package core

import (
	"github.com/lowe21/lxv/core/configcenter"
	"github.com/lowe21/lxv/core/logcenter"
	"github.com/lowe21/lxv/core/orm"
	"github.com/lowe21/lxv/core/swagger"
	"github.com/lowe21/lxv/core/trace"
)

func init() {
	configcenter.Init()
	logcenter.Init()
	orm.Init()
	swagger.Init()
	trace.Init()
}
