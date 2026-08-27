package filter

import (
	"context"

	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/filter"
	"dubbo.apache.org/dubbo-go/v3/protocol/base"
	"dubbo.apache.org/dubbo-go/v3/protocol/result"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

func init() {
	extension.SetFilter("trace-provider", func() filter.Filter {
		return &traceProviderFilter{}
	})
}

type traceProviderFilter struct{}

func (t *traceProviderFilter) Invoke(ctx context.Context, invoker base.Invoker, invocation base.Invocation) (res result.Result) {
	traceId, _ := invocation.GetAttachment("trace-id")
	if traceId == "" {
		return invoker.Invoke(ctx, invocation)
	}

	ctx, _ = gtrace.WithTraceID(ctx, traceId)
	ctx, span := otel.Tracer("dubbo.apache.org/dubbo-go/v3", trace.WithInstrumentationVersion(constant.Version)).
		Start(ctx, gstr.Join([]string{invoker.GetURL().Service(), invocation.MethodName()}, "."), trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(gtrace.CommonLabels()...))
	span.SetAttributes(
		attribute.String("dubbo.url", invoker.GetURL().String()),
	)
	span.AddEvent("dubbo.invoke", trace.WithAttributes(
		attribute.String("dubbo.invoke.arguments", gconv.String(invocation.Arguments())),
	))
	defer func() {
		span.AddEvent("dubbo.response", trace.WithAttributes(
			attribute.String("dubbo.response.result", gconv.String(res.Result())),
		))
		if err := res.Error(); err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	return invoker.Invoke(ctx, invocation)
}

func (t *traceProviderFilter) OnResponse(_ context.Context, result result.Result, _ base.Invoker, _ base.Invocation) result.Result {
	return result
}
