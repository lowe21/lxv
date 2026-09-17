package filter

import (
	"context"

	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/filter"
	"dubbo.apache.org/dubbo-go/v3/protocol/base"
	"dubbo.apache.org/dubbo-go/v3/protocol/result"

	"github.com/apache/dubbo-go-hessian2/java_exception"

	sentinelapi "github.com/alibaba/sentinel-golang/api"
	sentinelbase "github.com/alibaba/sentinel-golang/core/base"

	"github.com/lowe21/lxv/pkg/errcode"
)

func init() {
	extension.SetFilter("sentinel", func() filter.Filter {
		return &sentinel{}
	})
}

type sentinel struct{}

func (s *sentinel) Invoke(ctx context.Context, invoker base.Invoker, invocation base.Invocation) (res result.Result) {
	entry, block := sentinelapi.Entry(invoker.GetURL().Service()+"."+invocation.MethodName(), sentinelapi.WithResourceType(sentinelbase.ResTypeRPC), sentinelapi.WithTrafficType(sentinelbase.Inbound))
	if block != nil {
		subCode, message := errcode.Parse(errcode.ErrSystemBusy)

		return &result.RPCResult{
			Err: java_exception.NewThrowable(subCode + "@" + message),
		}
	}
	defer func() {
		if res != nil {
			if err := res.Error(); err != nil {
				sentinelapi.TraceError(entry, err)
			}
		}
		entry.Exit()
	}()

	return invoker.Invoke(ctx, invocation)
}

func (s *sentinel) OnResponse(_ context.Context, result result.Result, _ base.Invoker, _ base.Invocation) result.Result {
	return result
}
