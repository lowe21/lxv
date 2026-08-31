package client

import (
	"context"

	"github.com/lowe21/lxv/pkg/errcode"
)

type (
	OrderCancelReq struct {
		OrderCode string `json:"orderCode"`
	}
	OrderCancelRes struct{}
)

func OrderCancel(ctx context.Context, req *OrderCancelReq) (res *OrderCancelRes, err error) {
	defer func() {
		if err != nil {
			err = errcode.New(err)
		}
	}()

	return order.OrderCancel(ctx, req)
}
