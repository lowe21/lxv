package client

import (
	"context"

	"github.com/lowe21/lxv/pkg/dubbo"
)

func init() {
	dubbo.SetClient(order, &dubbo.ClientInfo{
		InterfaceName: order.Reference(),
		ConnectionInjectFunc: func(raw any, conn *dubbo.ClientConn) {
			client := raw.(*Order)
			client.orderCancel = func(ctx context.Context, req *OrderCancelReq) (res *OrderCancelRes, err error) {
				err = conn.CallUnary(ctx, []any{req}, &res, "orderCancel", dubbo.WithRetries(0))
				return
			}
		},
	})
}

var order = &Order{}

type Order struct {
	orderCancel func(context.Context, *OrderCancelReq) (*OrderCancelRes, error)
}

func (o *Order) Reference() string {
	return "order"
}
