package factory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/gogf/gf/v2/crypto/gsha256"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/lowe21/lxv/pkg/errcode"
)

type Factory[T any] struct {
	providers map[string]Provider[T]
	cache     *gcache.Cache
	ttl       time.Duration
	sf        singleflight.Group
	mutex     sync.RWMutex
}

func (f *Factory[T]) GetProvider(name string) (provider Provider[T], err error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	provider, ok := f.providers[name]
	if !ok {
		err = errcode.New(fmt.Sprintf("provider not found, name: %s", name))
	}

	return
}

func (f *Factory[T]) SetProvider(provider Provider[T]) {
	name := provider.Name()
	if name == "" {
		panic(fmt.Sprintf("provider name is empty, type: %T", provider))
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.providers == nil {
		f.providers = make(map[string]Provider[T])
	}
	if _, ok := f.providers[name]; ok {
		panic(fmt.Sprintf("provider already exists, name: %s", name))
	}
	f.providers[name] = provider
}

func (f *Factory[T]) Instance(name string, overrides map[string]any) (instance T, err error) {
	provider, err := f.GetProvider(name)
	if err != nil {
		return
	}

	ctx := context.Background()
	options := provider.Options()
	if len(overrides) > 0 {
		if err = gconv.Scan(overrides, options); err != nil {
			return
		}
	}
	key := gstr.Join([]string{name, gsha256.Encrypt(gconv.String(options))}, ":")

	value, err := f.cache.Get(ctx, key)
	if err != nil {
		return
	}
	if value != nil {
		instance = value.Val().(T)
	} else {
		result := <-f.sf.DoChan(key, func() (instance any, err error) {
			instance, err = provider.New(options)
			if err != nil {
				return
			}
			if g.IsNil(instance) {
				err = errcode.New(fmt.Sprintf("provider instance is nil, name: %s", name))
			} else {
				err = f.cache.Set(ctx, key, instance, f.ttl)
			}
			return
		})
		if result.Err != nil {
			err = result.Err
		} else {
			instance = result.Val.(T)
		}
	}

	return
}
