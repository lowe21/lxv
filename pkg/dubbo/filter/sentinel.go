package filter

import (
	"context"

	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/filter"
	"dubbo.apache.org/dubbo-go/v3/protocol/base"
	"dubbo.apache.org/dubbo-go/v3/protocol/result"
	"github.com/apache/dubbo-go-hessian2/java_exception"

	sentinel "github.com/alibaba/sentinel-golang/api"
	constant "github.com/alibaba/sentinel-golang/core/base"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"

	"github.com/lowe21/lxv/util"
)

func init() {
	extension.SetFilter("sentinel", func() filter.Filter {
		return &sentinelFilter{}
	})
}

type sentinelFilter struct{}

func (s *sentinelFilter) Invoke(ctx context.Context, invoker base.Invoker, invocation base.Invocation) (res result.Result) {
	entry, block := sentinel.Entry(gstr.Join([]string{invoker.GetURL().Service(), invocation.MethodName()}, "."), sentinel.WithResourceType(constant.ResTypeRPC), sentinel.WithTrafficType(constant.Inbound))
	if block != nil {
		errorCode := gerror.Code(util.Error(gcode.CodeServerBusy)).(util.ErrorCode)

		return &result.RPCResult{
			Err: java_exception.NewThrowable(
				gstr.Join([]string{errorCode.SubCode(), errorCode.Message()}, "@"),
			),
		}
	}
	defer func() {
		if err := res.Error(); err != nil {
			sentinel.TraceError(entry, err)
		}
		entry.Exit()
	}()

	return invoker.Invoke(ctx, invocation)
}

func (s *sentinelFilter) OnResponse(_ context.Context, result result.Result, _ base.Invoker, _ base.Invocation) result.Result {
	return result
}
