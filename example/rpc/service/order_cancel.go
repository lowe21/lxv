package service

import (
	"context"

	"github.com/lowe21/lxv/pkg/errcode"
	"github.com/lowe21/lxv/pkg/validation"
)

type (
	OrderCancelReq struct {
		OrderCode string `json:"orderCode" valid:"required"`
	}
	OrderCancelRes struct{}
)

func (o *Order) OrderCancel(ctx context.Context, req *OrderCancelReq) (res *OrderCancelRes, err error) {
	defer func() {
		if exception := recover(); exception != nil {
			err = errcode.New(exception)
		}
	}()

	if err = validation.Validator(ctx, req); err != nil {
		return
	}

	return &OrderCancelRes{}, nil
}
