package filter

import (
	"context"

	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/filter"
	"dubbo.apache.org/dubbo-go/v3/protocol/base"
	"dubbo.apache.org/dubbo-go/v3/protocol/result"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gogf/gf/v2/text/gstr"

	"github.com/lowe21/lxv/pkg/errcode"
)

func init() {
	extension.SetFilter("status", func() filter.Filter {
		return &statusFilter{}
	})
}

type statusFilter struct{}

func (s *statusFilter) Invoke(ctx context.Context, invoker base.Invoker, invocation base.Invocation) result.Result {
	return invoker.Invoke(ctx, invocation)
}

func (s *statusFilter) OnResponse(_ context.Context, result result.Result, _ base.Invoker, _ base.Invocation) result.Result {
	if err := result.Error(); err != nil {
		subCode, message := errcode.Parse(errcode.New(err))

		result.SetError(
			status.Error(codes.Internal, gstr.Join([]string{subCode, message}, "@")),
		)
	}

	return result
}
