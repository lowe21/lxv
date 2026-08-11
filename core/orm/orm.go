package orm

import (
	"sync"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/util/gconv"
)

var once sync.Once

func Init() {
	once.Do(func() {
		database := g.Config().MustGet(nil, "database").Map()
		redis := g.Config().MustGet(nil, "redis").Map()
		for name := range database {
			if config, ok := redis[name]; ok {
				if err := gredis.SetConfigByMap(gconv.Map(config), name); err != nil {
					panic(err)
				}
				g.DB(name).GetCache().SetAdapter(gcache.NewAdapterRedis(g.Redis(name)))
			}
		}
	})
}
