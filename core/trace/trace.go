package trace

import (
	"sync"

	"github.com/gogf/gf/contrib/trace/otlpgrpc/v2"
	"github.com/gogf/gf/v2/frame/g"
)

var once sync.Once

func Init() {
	once.Do(func() {
		address := g.Config().MustGet(nil, "jaeger.address").String()
		token := g.Config().MustGet(nil, "jaeger.token").String()
		if address != "" {
			if _, err := otlpgrpc.Init(g.Server().GetName(), address, token); err != nil {
				panic(err)
			}
		}
	})
}
