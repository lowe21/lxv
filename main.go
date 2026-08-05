package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	_ "github.com/lowe21/lxv/core"
)

func main() {
	g.Dump(g.Config().Data(gctx.New()))
}
