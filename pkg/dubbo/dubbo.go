package dubbo

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"dubbo.apache.org/dubbo-go/v3"
	"dubbo.apache.org/dubbo-go/v3/client"
	"dubbo.apache.org/dubbo-go/v3/common"
	_ "dubbo.apache.org/dubbo-go/v3/imports"

	hessian2 "github.com/apache/dubbo-go-hessian2"

	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gutil"

	_ "github.com/lowe21/lxv/pkg/dubbo/filter"
)

type (
	ClientInfo    = client.ClientInfo
	ClientConn    = client.Connection
	ClientOptions = client.CallOptions
	ClientOption  = client.CallOption
)

func Load() {
	if err := dubbo.Load(); err != nil {
		panic(err)
	}
}

func SetService(service common.RPCService) {
	reference := common.GetReference(service)
	if reference == "" {
		panic(fmt.Sprintf("service reference is empty, type: %T", service))
	}

	registerPOJOs(reference, getRPCPOJOTypes(service))
	dubbo.SetProviderServiceWithInfo(service, &common.ServiceInfo{
		InterfaceName: reference,
	})
}

func SetClient(service common.RPCService, info *ClientInfo) {
	reference := common.GetReference(service)
	if reference == "" {
		panic(fmt.Sprintf("client reference is empty, type: %T", service))
	}

	registerPOJOs(reference, getRPCPOJOTypes(service))
	dubbo.SetConsumerServiceWithInfo(service, info)
}

func WithRequestTimeout(timeout time.Duration) ClientOption {
	return func(opts *ClientOptions) {
		opts.RequestTimeout = timeout.String()
	}
}

func WithRetries(retries int) ClientOption {
	return func(opts *ClientOptions) {
		opts.Retries = gconv.String(retries)
	}
}

func registerPOJOs(reference string, pojoTypes []reflect.Type) {
	var (
		register func(reflect.Type)
		visited  = make(map[reflect.Type]struct{})
	)

	register = func(elem reflect.Type) {
		if _, ok := visited[elem]; ok {
			return
		}
		visited[elem] = struct{}{}

		for _, subElem := range getFieldPOJOTypes(elem) {
			register(subElem)
		}

		object := reflect.New(elem).Interface()

		name := elem.Name()
		if pojo, ok := object.(hessian2.POJO); ok {
			if javaClassName := pojo.JavaClassName(); javaClassName != "" {
				name = javaClassName
			}
		}

		hessian2.RegisterPOJOMapping(gstr.CaseDelimitedScreaming(
			reference+"."+name,
			'.',
			false,
		), object)
	}

	for _, pojoType := range pojoTypes {
		register(pojoType)
	}
}

func getRPCPOJOTypes(service common.RPCService) (pojoTypes []reflect.Type) {
	reflectValue := gutil.OriginValueAndKind(service)
	functions := make([]reflect.Type, 0)

	if reflectValue.OriginValue.NumField() > 0 {
		for _, field := range reflectValue.OriginValue.Fields() {
			if field.Kind() != reflect.Func {
				panic(fmt.Sprintf("invalid field %v, must be function, type: %T", field.Type(), service))
			}
			functions = append(functions, field.Type())
		}
	} else {
		for method, value := range reflectValue.InputValue.Methods() {
			if method.Name != "Reference" {
				functions = append(functions, value.Type())
			}
		}
	}

	if len(functions) == 0 {
		panic(fmt.Sprintf("%T has no callable methods or function fields", service))
	}

	contextType := reflect.TypeFor[context.Context]()
	errorType := reflect.TypeFor[error]()

	for _, function := range functions {
		if function.NumIn() != 2 || function.NumOut() != 2 {
			panic(fmt.Sprintf("invalid function %v, required func(context.Context, *XxReq) (*XxRes, error), type: %T", function, service))
		}

		if function.In(0) != contextType {
			panic(fmt.Sprintf("invalid function %v, first input must be context.Context, type: %T", function, service))
		}

		reqType := function.In(1)
		if reqType.Kind() != reflect.Pointer || reqType.Elem().Kind() != reflect.Struct {
			panic(fmt.Sprintf("invalid function %v, second input must be pointer to struct like *XxReq, type: %T", function, service))
		}

		resType := function.Out(0)
		if resType.Kind() != reflect.Pointer || resType.Elem().Kind() != reflect.Struct {
			panic(fmt.Sprintf("invalid function %v, first output must be pointer to struct like *XxRes, type: %T", function, service))
		}

		if function.Out(1) != errorType {
			panic(fmt.Sprintf("invalid function %v, second output must be error, type: %T", function, service))
		}

		pojoTypes = append(pojoTypes, reqType.Elem(), resType.Elem())
	}

	return
}

func getFieldPOJOTypes(structType reflect.Type) (pojoTypes []reflect.Type) {
	var (
		handler func(reflect.Type)
		visited = make(map[reflect.Type]struct{})
		pkgPath = structType.PkgPath()
	)

	handler = func(fieldType reflect.Type) {
		if _, ok := visited[fieldType]; ok {
			return
		}
		visited[fieldType] = struct{}{}

		switch fieldType.Kind() {
		case reflect.Struct:
			if fieldType.PkgPath() == pkgPath {
				panic(fmt.Sprintf("invalid struct %v, must be pointer to struct", fieldType))
			}
		case reflect.Pointer:
			elem := fieldType.Elem()
			if elem.Kind() == reflect.Struct && elem.PkgPath() == pkgPath {
				pojoTypes = append(pojoTypes, elem)
			}
		case reflect.Slice, reflect.Array:
			handler(fieldType.Elem())
		case reflect.Map:
			handler(fieldType.Key())
			handler(fieldType.Elem())
		default:
			return
		}
	}

	for field := range structType.Fields() {
		if field.PkgPath != "" || field.Tag.Get("hessian") == "-" {
			continue
		}
		handler(field.Type)
	}

	return
}
