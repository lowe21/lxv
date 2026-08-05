package log_center

import (
	"context"
	"fmt"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/lowe21/lxv/core/log_center/graylog"
)

var once sync.Once

func Init() {
	once.Do(func() {
		if err := g.Log().SetConfig(g.Server().Logger().GetConfig()); err != nil {
			panic(err)
		}

		logCenter := graylog.New()

		g.Log().SetHandlers(func(ctx context.Context, input *glog.HandlerInput) {
			input.Next(ctx)

			content := input.ValuesContent()
			delimiter := "\n"
			stack := ""

			for _, value := range input.Values {
				strings := fmt.Sprintf("%+v", value)
				if strings != "" {
					array := gstr.Split(strings, delimiter)
					temp := make([]string, 0, len(array))
					for _, item := range array {
						if item != content {
							temp = append(temp, item)
						}
					}
					if len(temp) > 0 {
						stack = gstr.Join([]string{stack, gstr.Join(temp, delimiter)}, delimiter)
					}
				}
			}

			if stack == "" {
				stack = input.Stack
			}

			logCenter.Send(ctx, &graylog.Gelf{
				Version:      "1.1",
				Host:         g.Server().GetName(),
				ShortMessage: content,
				FullMessage:  stack,
				Timestamp:    gconv.Float64(input.Time.UnixMilli()) / 1e3,
				Level:        input.Level,
				LevelFormat:  input.LevelFormat,
			})
		})
	})
}
