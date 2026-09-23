package logcenter

import (
	"context"
	"fmt"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"

	"github.com/lowe21/lxv/pkg/graylog"
)

var once sync.Once

func Init() {
	once.Do(func() {
		if err := g.Log().SetConfig(g.Server().Logger().GetConfig()); err != nil {
			panic(err)
		}

		if g.Config().MustGet(nil, "graylog.address").String() != "" {
			g.Log().SetHandlers(func(ctx context.Context, input *glog.HandlerInput) {
				input.Next(ctx)

				content := input.ValuesContent()
				stacks := make([]string, 1, len(input.Values)+1)
				delimiter := "\n"

				for _, value := range input.Values {
					if fmtStr := fmt.Sprintf("%+v", value); fmtStr != "" {
						for _, str := range gstr.Split(fmtStr, delimiter) {
							if str != content {
								stacks = append(stacks, str)
							}
						}
					}
				}

				stack := gstr.Join(stacks, delimiter)
				if stack == "" {
					stack = input.Stack
				}

				graylog.Send(&graylog.Gelf{
					Version:      graylog.GelfVersion,
					Host:         g.Server().GetName(),
					ShortMessage: content,
					FullMessage:  stack,
					Timestamp:    float64(input.Time.UnixMilli()) / 1e3,
					Level:        input.Level,
					LevelFormat:  input.LevelFormat,
				})
			})
		}
	})
}
