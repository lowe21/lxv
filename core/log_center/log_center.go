package log_center

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
				delimiter := "\n"
				stack := ""

				for _, value := range input.Values {
					strings := fmt.Sprintf("%+v", value)
					if strings != "" {
						array := gstr.Split(strings, delimiter)
						stacks := make([]string, 0, len(array))
						for _, item := range array {
							if item != content {
								stacks = append(stacks, item)
							}
						}
						if len(stacks) > 0 {
							stack = gstr.Join([]string{stack, gstr.Join(stacks, delimiter)}, delimiter)
						}
					}
				}

				if stack == "" {
					stack = input.Stack
				}

				graylog.Send(&graylog.Gelf{
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
