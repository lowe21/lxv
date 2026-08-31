package service

import (
	"github.com/lowe21/lxv/pkg/dubbo"
)

func init() {
	dubbo.SetService(&Order{})
}

type Order struct{}

func (o *Order) Reference() string {
	return "order"
}
