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
	options      *Options
	mutex        *redsync.Mutex
	locked       atomic.Bool
	extendCancel context.CancelFunc
}

func (r *RedMutex) Lock(ctx context.Context) (err error) {
	if r.locked.Load() {
		return redsync.ErrFailed
	}

	lockCtx := ctx
	lockCancel := context.CancelFunc(nil)

	if r.options.LockTimeout > 0 {
		lockCtx, lockCancel = context.WithTimeout(ctx, r.options.LockTimeout)
		defer lockCancel()
	}

	if err = r.mutex.LockContext(lockCtx); err == nil {
		r.locked.Store(true)
		r.extend(ctx)
	}

	return
}

func (r *RedMutex) TryLock(ctx context.Context) (err error) {
	if r.locked.Load() {
		return redsync.ErrFailed
	}

	if err = r.mutex.TryLockContext(ctx); err == nil {
		r.locked.Store(true)
		r.extend(ctx)
	}

	return
}

func (r *RedMutex) Unlock(ctx context.Context) (err error) {
	if r.extendCancel != nil {
		r.extendCancel()
		r.extendCancel = nil
	}

	unlockCtx, unlockCancel := context.WithTimeout(ctx, r.options.UnlockTimeout)
	defer func() {
		r.locked.Store(false)
		unlockCancel()
	}()

	if _, err = r.mutex.UnlockContext(unlockCtx); err != nil {
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
	interval := r.options.Expiry / 3

	extendCtx, extendCancel := context.WithCancel(ctx)
	r.extendCancel = extendCancel

	go func() {
		ticker := time.NewTicker(interval)
		defer func() {
			ticker.Stop()
			extendCancel()
		}()

		for {
			select {
			case <-ticker.C:
				if _, err := r.mutex.ExtendContext(extendCtx); err != nil {
					if extendCtx.Err() != nil {
						return
					}
					g.Log().Error(ctx, err)
				}
			case <-extendCtx.Done():
				return
			}
		}
	}()
}
