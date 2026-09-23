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
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/lowe21/lxv/pkg/errcode"
)

type Factory[T any] struct {
	providers map[string]providerEntry[T]
	cache     *gcache.Cache
	ttl       time.Duration
	mutex     sync.RWMutex
	sf        singleflight.Group
}

func (f *Factory[T]) SetProvider[O any](provider Provider[T, O]) {
	name := provider.Name()
	if name == "" {
		panic(fmt.Sprintf("provider name is empty, type: %T", provider))
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if _, ok := f.providers[name]; ok {
		panic(fmt.Sprintf("provider already exists, name: %s", name))
	}
	f.providers[name] = &providerAdapter[T, O]{
		provider: provider,
	}
}

func (f *Factory[T]) Instance(name string, overrides map[string]any) (instance T, err error) {
	f.mutex.RLock()
	provider, ok := f.providers[name]
	f.mutex.RUnlock()
	if !ok {
		err = errcode.New(fmt.Sprintf("provider not found, name: %s", name))
		return
	}

	options := provider.Options()
	if g.IsNil(options) {
		err = errcode.New(fmt.Sprintf("provider options is nil, name: %s", name))
		return
	}
	if len(overrides) > 0 {
		if err = gconv.Scan(overrides, options); err != nil {
			return
		}
	}
	key := name + ":" + gsha256.Encrypt(gconv.String(options))

	instance = f.getInstance(key)
	if !g.IsNil(instance) {
		return
	}

	value, err, _ := f.sf.Do(key, func() (value any, err error) {
		cachedInstance := f.getInstance(key)
		if !g.IsNil(cachedInstance) {
			return cachedInstance, nil
		}

		newInstance, err := provider.New(options)
		if err != nil {
			return
		}
		if g.IsNil(newInstance) {
			err = errcode.New(fmt.Sprintf("provider instance is nil, name: %s", name))
			return
		}
		f.setInstance(key, newInstance)

		return newInstance, nil
	})
	if err != nil {
		return
	}

	return value.(T), nil
}

func (f *Factory[T]) getInstance(key string) (instance T) {
	value, err := f.cache.Get(context.Background(), key)
	if err != nil {
		return
	}
	if value != nil {
		instance = value.Val().(T)
	}

	return
}

func (f *Factory[T]) setInstance(key string, instance T) {
	_ = f.cache.Set(context.Background(), key, instance, f.ttl)
}
