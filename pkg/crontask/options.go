package crontask

import (
	"slices"

	"github.com/gogf/gf/v2/frame/g"
)

type Option struct {
	Name    string
	Pattern string
}

func defaultOptions() []*Option {
	options := make([]*Option, 0)
	if err := g.Config().MustGet(nil, "crontask").Scan(&options); err != nil {
		panic(err)
	}

	return slices.DeleteFunc(options, func(option *Option) bool {
		return option.Name == "" || option.Pattern == ""
	})
}
