package filter

import (
	"context"

	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/filter"
	"dubbo.apache.org/dubbo-go/v3/protocol/base"
	"dubbo.apache.org/dubbo-go/v3/protocol/result"

	sentinel "github.com/alibaba/sentinel-golang/api"
	constant "github.com/alibaba/sentinel-golang/core/base"

	"github.com/apache/dubbo-go-hessian2/java_exception"

	"github.com/gogf/gf/v2/text/gstr"

	"github.com/lowe21/lxv/pkg/errcode"
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
		subCode, message := errcode.Parse(errcode.SystemBusy)

		return &result.RPCResult{
			Err: java_exception.NewThrowable(gstr.Join([]string{subCode, message}, "@")),
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
