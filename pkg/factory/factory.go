package factory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"

	"github.com/lowe21/lxv/pkg/errcode"
)

type Factory[T any] struct {
	providers map[string]Provider[T]
	instances *gcache.Cache
	ttl       time.Duration
	mutex     sync.RWMutex
	sf        singleflight.Group
}

func (f *Factory[T]) Register(name string, provider Provider[T]) {
	if name == "" {
		panic(fmt.Sprintf("provider name is empty, type: %T", provider))
	}
	if g.IsNil(provider) {
		panic(fmt.Sprintf("provider is nil, name: %s", name))
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if _, ok := f.providers[name]; ok {
		panic(fmt.Sprintf("provider already exists, name: %s", name))
	}
	f.providers[name] = provider
}

func (f *Factory[T]) Instance(name string, overrides map[string]any) (instance T, err error) {
	f.mutex.RLock()
	provider, ok := f.providers[name]
	f.mutex.RUnlock()
	if !ok {
		err = errcode.New(errcode.ErrBusinessFailed, fmt.Sprintf("provider is not registered, name: %s", name))
		return
	}

	instanceKey, optionsMap, err := provider.Prepare(name, overrides)
	if err != nil {
		return
	}

	ctx := context.Background()

	value, err := f.instances.Get(ctx, instanceKey)
	if err != nil {
		return
	}
	if value != nil {
		return value.Val().(T), nil
	}

	result := <-f.sf.DoChan(instanceKey, func() (instance any, err error) {
		instance, err = provider.New(optionsMap)
		if err != nil {
			return
		}
		if !g.IsNil(instance) {
			err = f.instances.Set(ctx, instanceKey, instance, f.ttl)
		} else {
			err = errcode.New(errcode.ErrBusinessFailed, fmt.Sprintf("provider instance is nil, name: %s", name))
		}
		return
	})
	if result.Err != nil {
		err = result.Err
	} else {
		instance = result.Val.(T)
	}

	return
}
