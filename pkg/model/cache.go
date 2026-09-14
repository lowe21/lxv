package model

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

type cacheInvalidator struct {
	keys  map[string]map[string]struct{}
	mutex sync.RWMutex
}

func (c *cacheInvalidator) register(group, key string) {
	if key == "" {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.keys == nil {
		c.keys = make(map[string]map[string]struct{})
	}
	if c.keys[group] == nil {
		c.keys[group] = make(map[string]struct{})
	}
	c.keys[group][key] = struct{}{}
}

func (c *cacheInvalidator) flush(ctx context.Context, db gdb.DB) {
	group := db.GetGroup()

	c.mutex.RLock()
	keys := make([]string, 0, len(c.keys[group]))
	for key := range c.keys[group] {
		keys = append(keys, key)
	}
	c.mutex.RUnlock()

	if len(keys) > 0 {
		ctx = context.WithoutCancel(ctx)

		if err := db.GetCache().Removes(ctx, gconv.SliceAny(keys)); err != nil {
			g.Log().Errorf(ctx, "flush cache error, group: %s, keys: %v, %v", group, keys, err)
		} else {
			c.mutex.Lock()
			for _, key := range keys {
				delete(c.keys[group], key)
			}
			if len(c.keys[group]) == 0 {
				delete(c.keys, group)
			}
			c.mutex.Unlock()
		}
	}
}

type cacheInvalidatorCtxKey struct{}

var invalidatorCtxKey = cacheInvalidatorCtxKey{}

func cacheInvalidatorFromCtx(ctx context.Context) (invalidator *cacheInvalidator) {
	if ctx == nil {
		return
	}

	invalidator, ok := ctx.Value(invalidatorCtxKey).(*cacheInvalidator)
	if !ok {
		return
	}

	return
}

func cacheInvalidate(ctx context.Context, db gdb.DB, keys ...string) {
	set := gset.NewStrSet()
	for _, key := range keys {
		if key == "" {
			continue
		}
		if !gstr.HasPrefix(key, "SelectCache:") {
			key = gstr.Join([]string{"SelectCache", key}, ":")
		}
		set.Add(key)
	}
	if set.Size() == 0 {
		return
	}

	keys = set.Slice()
	group := db.GetGroup()

	if gdb.TXFromCtx(ctx, group) != nil {
		if invalidator := cacheInvalidatorFromCtx(ctx); invalidator != nil {
			for _, key := range keys {
				invalidator.register(group, key)
			}
		} else {
			g.Log().Errorf(ctx, "transaction context missing cache invalidator, group: %s, keys: %v", group, keys)
		}
		return
	}

	ctx = context.WithoutCancel(ctx)

	if err := db.GetCache().Removes(ctx, gconv.SliceAny(keys)); err != nil {
		g.Log().Errorf(ctx, "flush cache error, group: %s, keys: %v, %v", group, keys, err)
	}
}

func cacheHandler(db gdb.DB, keys ...string) (handler gdb.HookHandler) {
	return gdb.HookHandler{
		Insert: func(ctx context.Context, input *gdb.HookInsertInput) (result sql.Result, err error) {
			defer func() {
				if err == nil {
					cacheInvalidate(ctx, db, keys...)
				}
			}()

			return input.Next(ctx)
		},
		Update: func(ctx context.Context, input *gdb.HookUpdateInput) (result sql.Result, err error) {
			defer func() {
				if err == nil {
					cacheInvalidate(ctx, db, keys...)
				}
			}()

			return input.Next(ctx)
		},
		Delete: func(ctx context.Context, input *gdb.HookDeleteInput) (result sql.Result, err error) {
			defer func() {
				if err == nil {
					cacheInvalidate(ctx, db, keys...)
				}
			}()

			return input.Next(ctx)
		},
	}
}

func cacheOption(db gdb.DB, table, key string, ttl ...time.Duration) (option gdb.CacheOption) {
	duration := time.Duration(0)
	if len(ttl) > 0 {
		duration = ttl[0]
	}

	if key != "" {
		key = gstr.Join([]string{":", key}, "")
	}

	return gdb.CacheOption{
		Duration: duration,
		Name:     gstr.Join([]string{db.GetGroup(), "@", db.GetSchema(), "#", table, key}, ""),
		Force:    true,
	}
}
