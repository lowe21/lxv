package redsync

import (
	"context"

	"github.com/go-redsync/redsync/v4"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/gogf/gf/v2/util/gconv"
)

type RedMutex struct {
	options      *Options
	mutex        *redsync.Mutex
	entry        *gtimer.Entry
	extendCancel context.CancelFunc
}

func (r *RedMutex) Lock(ctx context.Context) (err error) {
	if err = r.mutex.LockContext(ctx); err != nil {
		return
	}

	r.extend(ctx)
	return
}

func (r *RedMutex) TryLock(ctx context.Context) (err error) {
	if err = r.mutex.TryLockContext(ctx); err != nil {
		return
	}

	r.extend(ctx)
	return
}

func (r *RedMutex) Unlock(ctx context.Context) (err error) {
	if r.extendCancel != nil {
		r.extendCancel()
	}

	if r.entry != nil {
		r.entry.Close()
	}

	if _, err = r.mutex.UnlockContext(ctx); err != nil {
		if gerror.Is(err, redsync.ErrLockAlreadyExpired) {
			err = nil
		} else {
			g.Log().Error(ctx, err)
		}
		return
	}

	return
}

func (r *RedMutex) extend(ctx context.Context) {
	interval := r.options.Expiry / 3
	if interval <= 0 {
		interval = gtime.S
	}

	times := gconv.Int(r.options.ExtendDuration.Seconds() / interval.Seconds())
	if times <= 0 {
		times = 1
	}

	extendCtx, cancel := context.WithCancel(ctx)
	r.extendCancel = cancel

	r.entry = gtimer.AddTimes(extendCtx, interval, times, func(ctx context.Context) {
		if _, err := r.mutex.ExtendContext(ctx); err != nil {
			g.Log().Error(ctx, err)
		}
	})

	go func() {
		<-extendCtx.Done()
		if r.entry != nil {
			r.entry.Close()
		}
	}()
}
