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

	registerPOJO(reference, getPOJOElems(service))
	dubbo.SetProviderServiceWithInfo(service, &common.ServiceInfo{
		InterfaceName: reference,
	})
}

func SetClient(service common.RPCService, info *ClientInfo) {
	reference := common.GetReference(service)
	if reference == "" {
		panic(fmt.Sprintf("client reference is empty, type: %T", service))
	}

	registerPOJO(reference, getPOJOElems(service))
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

func registerPOJO(reference string, elems []reflect.Type) {
	for _, elem := range elems {
		object := reflect.New(elem).Interface()
		reflectValue := gutil.OriginValueAndKind(object)
		reflectType := gutil.OriginTypeAndKind(object)

		if reflectValue.OriginValue.NumField() > 0 {
			pkgPath := reflectType.OriginType.PkgPath()
			subElems := make([]reflect.Type, 0)

			for i := 0; i < reflectValue.OriginValue.NumField(); i++ {
				fieldType := reflectValue.OriginValue.Field(i).Type()
				if gutil.OriginTypeAndKind(fieldType).OriginType.PkgPath() != pkgPath {
					continue
				}
				if fieldType.Kind() != reflect.Ptr || (fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() != reflect.Struct) {
					panic(fmt.Sprintf("invalid struct defined as %v, but the parameter should be a pointer to struct", fieldType))
				}
				subElems = append(subElems, fieldType.Elem())
			}

			if len(subElems) > 0 {
				registerPOJO(reference, subElems)
			}
		}

		name := gstr.StrEx(elem.String(), ".")
		if gstr.PosI(name, reference) == 0 {
			name = gstr.SubStr(name, len(reference))
		}
		if pojo, ok := object.(hessian2.POJO); ok {
			if javaClassName := pojo.JavaClassName(); javaClassName != "" {
				name = javaClassName
			}
		}

		hessian2.RegisterPOJOMapping(gstr.CaseDelimitedScreaming(gstr.Join([]string{reference, name}, "."), byte('.'), false), object)
	}
}

func getPOJOElems(service common.RPCService) (elems []reflect.Type) {
	reflectValue := gutil.OriginValueAndKind(service)
	reflectType := gutil.OriginTypeAndKind(service)
	functions := make([]reflect.Type, 0)

	if reflectValue.OriginValue.NumField() > 0 {
		for i := 0; i < reflectValue.OriginValue.NumField(); i++ {
			field := reflectValue.OriginValue.Field(i)
			switch field.Kind() {
			case reflect.Func:
				functions = append(functions, field.Type())
			default:
				panic(fmt.Sprintf("invalid function parameter type %v", field.Type()))
			}
		}
	} else {
		for i := 0; i < reflectValue.InputValue.NumMethod(); i++ {
			if reflectType.InputType.Method(i).Name != "Reference" {
				functions = append(functions, reflectValue.InputValue.Method(i).Type())
			}
		}
	}

	for _, function := range functions {
		numIn := function.NumIn()
		numOut := function.NumOut()
		if numIn != 2 || numOut != 2 {
			panic(fmt.Sprintf("invalid function defined as %v, but func(context.Context, *XxReq) (*XxRes, error) is required", function))
		}

		for i := 0; i < numIn; i++ {
			in := function.In(i)
			if i != numIn-1 {
				if !in.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
					panic(fmt.Sprintf("invalid function defined as %v, but the first input parameter should be type of context.Context", function))
				}
			} else {
				if in.Kind() != reflect.Ptr || (in.Kind() == reflect.Ptr && in.Elem().Kind() != reflect.Struct) {
					panic(fmt.Sprintf("invalid function defined as %v, but the second input parameter should be type of pointer to struct like *XxReq", function))
				}
				elems = append(elems, in.Elem())
			}
		}

		for i := 0; i < numOut; i++ {
			out := function.Out(i)
			if i != numOut-1 {
				if out.Kind() != reflect.Ptr || (out.Kind() == reflect.Ptr && out.Elem().Kind() != reflect.Struct) {
					panic(fmt.Sprintf("invalid function defined as %v, but the first output parameter should be type of pointer to struct like *XxRes", function))
				}
				elems = append(elems, out.Elem())
			} else {
				if !out.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
					panic(fmt.Sprintf("invalid function defined as %v, but the second output parameter should be type of error", function))
				}
			}
		}
	}

	return
}
