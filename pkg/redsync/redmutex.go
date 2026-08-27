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
		lockCtx    = ctx
		lockCancel context.CancelFunc
	)

	if r.options.LockTimeout > 0 {
		lockCtx, lockCancel = context.WithTimeout(ctx, r.options.LockTimeout)
		defer lockCancel()
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
	extendCtx, extendCancel := context.WithCancel(ctx)
	r.cancel = extendCancel

	interval := r.options.Expiry / 3
	if interval <= 0 {
		interval = time.Second
	}

	extendCount := 0
	extendMax := 0
	if r.options.ExtendMaxDuration > 0 {
		extendMax = int(r.options.ExtendMaxDuration.Seconds() / interval.Seconds())
		if extendMax <= 0 {
			extendMax = 1
		}
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if extendMax > 0 && extendCount >= extendMax {
					return
				}
				if _, err := r.mutex.ExtendContext(ctx); err != nil {
					if extendCtx.Err() != nil {
						return
					}
					g.Log().Error(ctx, err)
				}
				extendCount++
			case <-extendCtx.Done():
				return
			}
		}
	}()
}
