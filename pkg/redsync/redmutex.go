package redsync

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/go-redsync/redsync/v4"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type RedMutex struct {
	options *Options
	mutex   *redsync.Mutex
	locked  atomic.Bool
	cancel  context.CancelFunc
}

func (r *RedMutex) Lock(ctx context.Context) (err error) {
	if r.locked.Load() {
		return redsync.ErrFailed
	}

	var (
		lockCtx = ctx
		cancel  context.CancelFunc
	)

	if r.options.LockTimeout > 0 {
		lockCtx, cancel = context.WithTimeout(ctx, r.options.LockTimeout)
		defer cancel()
	}

	if err = r.mutex.LockContext(lockCtx); err != nil {
		return
	}

	r.locked.Store(true)
	r.extend(ctx)

	return
}

func (r *RedMutex) TryLock(ctx context.Context) (err error) {
	if r.locked.Load() {
		return redsync.ErrFailed
	}

	if err = r.mutex.TryLockContext(ctx); err != nil {
		return
	}

	r.locked.Store(true)
	r.extend(ctx)

	return
}

func (r *RedMutex) Unlock(ctx context.Context) (err error) {
	if r.cancel != nil {
		r.cancel()
	}

	unLockCtx, cancel := context.WithTimeout(ctx, r.options.UnLockTimeout)
	defer func() {
		r.locked.Store(false)
		cancel()
	}()

	if _, err = r.mutex.UnlockContext(unLockCtx); err != nil {
		if gerror.Is(err, redsync.ErrLockAlreadyExpired) {
			err = nil
		} else if _, ok := errors.AsType[*redsync.ErrTaken](err); ok {
			err = nil
		} else if _, ok = errors.AsType[*redsync.ErrNodeTaken](err); ok {
			err = nil
		}
	}

	return
}

func (r *RedMutex) extend(ctx context.Context) {
	extendCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	interval := r.options.Expiry / 3
	if interval <= 0 {
		interval = time.Second
	}

	keepCount := 0
	totalCount := 0
	if r.options.ExtendDuration > 0 {
		totalCount = int(r.options.ExtendDuration.Seconds() / interval.Seconds())
		if totalCount <= 0 {
			totalCount = 1
		}
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if totalCount > 0 && keepCount >= totalCount {
					return
				}
				if _, err := r.mutex.ExtendContext(ctx); err != nil {
					if extendCtx.Err() != nil {
						return
					}
					g.Log().Error(ctx, err)
				}
				keepCount++
			case <-extendCtx.Done():
				return
			}
		}
	}()
}
