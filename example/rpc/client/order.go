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
			client.OrderCancel = func(ctx context.Context, req *OrderCancelReq) (res *OrderCancelRes, err error) {
				if err = conn.CallUnary(ctx, []any{req}, &res, "orderCancel", dubbo.WithRetries(0)); err != nil {
					return
				}
				return
			}
		},
	})
}

var order = &Order{}

type Order struct {
	OrderCancel func(ctx context.Context, req *OrderCancelReq) (res *OrderCancelRes, err error)
}

func (o *Order) Reference() string {
	return "order"
}
