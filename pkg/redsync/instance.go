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
	mutex   sync.Mutex
)

func instance() *RedSync {
	mutex.Lock()
	defer mutex.Unlock()

	if redSync == nil {
		options := newOptions()
		redSync = &RedSync{
			options: options,
			sync: redsync.New(goredis.NewPool(
				g.Redis(options.RedisGroup).Client().(redis.UniversalClient),
			)),
		}
	}

	return redSync
}

func Mutex(name string, opts ...Option) *RedMutex {
	return instance().Mutex(name, opts...)
}
