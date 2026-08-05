package redsync

import (
	"sync"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"

	"github.com/redis/go-redis/v9"

	"github.com/gogf/gf/v2/frame/g"
)

var (
	redSync *RedSync
	once    sync.Once
)

func instance() *RedSync {
	once.Do(func() {
		redSync = &RedSync{
			options: defaultOptions(),
			sync: redsync.New(
				goredis.NewPool(
					g.Redis().Client().(redis.UniversalClient),
				),
			),
		}
	})

	return redSync
}

func Mutex(name string, opts ...Option) *RedMutex {
	return instance().Mutex(name, opts...)
}
