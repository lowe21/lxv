package filter

import (
	"context"

	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/filter"
	"dubbo.apache.org/dubbo-go/v3/protocol/base"
	"dubbo.apache.org/dubbo-go/v3/protocol/result"

	"github.com/gogf/gf/v2/net/gtrace"
)

func init() {
	extension.SetFilter("trace-consumer", func() filter.Filter {
		return &traceConsumerFilter{}
	})
}

type traceConsumerFilter struct{}

func (t *traceConsumerFilter) Invoke(ctx context.Context, invoker base.Invoker, invocation base.Invocation) result.Result {
	invocation.SetAttachment("trace-id", gtrace.GetTraceID(ctx))

	return invoker.Invoke(ctx, invocation)
}

func (t *traceConsumerFilter) OnResponse(_ context.Context, result result.Result, _ base.Invoker, _ base.Invocation) result.Result {
	return result
}
