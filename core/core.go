package core

import (
	"github.com/lowe21/lxv/core/config_center"
	"github.com/lowe21/lxv/core/log_center"
	"github.com/lowe21/lxv/core/orm"
	"github.com/lowe21/lxv/core/swagger"
	"github.com/lowe21/lxv/core/trace"
)

func init() {
	config_center.Init()
	log_center.Init()
	orm.Init()
	swagger.Init()
	trace.Init()
}
