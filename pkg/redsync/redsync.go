package redsync

import (
	"github.com/go-redsync/redsync/v4"
)

type RedSync struct {
	options *Options
	sync    *redsync.Redsync
}

func (r *RedSync) Mutex(name string, opts ...Option) *RedMutex {
	options := *r.options
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return &RedMutex{
		options: &options,
		mutex: r.sync.NewMutex(name,
			redsync.WithExpiry(options.Expiry),
			redsync.WithTries(options.Tries),
			redsync.WithRetryDelay(options.RetryDelay),
			redsync.WithShufflePools(true),
		),
	}
}
