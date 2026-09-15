package factory

import (
	"time"

	"github.com/gogf/gf/v2/os/gcache"
)

func New[T any]() *Factory[T] {
	return &Factory[T]{
		providers: make(map[string]Provider[T]),
		instances: gcache.New(),
		ttl:       time.Hour,
	}
}
