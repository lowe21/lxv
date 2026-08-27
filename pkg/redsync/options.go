package redsync

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	defaultRedisGroup        = "default"
	defaultRedisKeyPrefix    = "lock"
	defaultExpiry            = "10s"
	defaultTries             = 60
	defaultRetryDelay        = "1s"
	defaultExtendMaxDuration = "0"
	defaultLockTimeout       = "3s"
	defaultUnlockTimeout     = "3s"
)

type Options struct {
	RedisGroup        string
	RedisKeyPrefix    string
	Expiry            time.Duration
	Tries             int
	RetryDelay        time.Duration
	ExtendMaxDuration time.Duration
	LockTimeout       time.Duration
	UnlockTimeout     time.Duration
}

func defaultOptions() *Options {
	options := &Options{}
	if err := g.Config().MustGet(nil, "redsync").Scan(options); err != nil {
		panic(err)
	}

	if options.RedisGroup == "" {
		options.RedisGroup = defaultRedisGroup
	}
	if options.RedisKeyPrefix == "" {
		options.RedisKeyPrefix = defaultRedisKeyPrefix
	}
	if options.Expiry <= 0 {
		options.Expiry = gconv.Duration(defaultExpiry)
	}
	if options.Tries <= 0 {
		options.Tries = defaultTries
	}
	if options.RetryDelay <= 0 {
		options.RetryDelay = gconv.Duration(defaultRetryDelay)
	}
	if options.ExtendMaxDuration < 0 {
		options.ExtendMaxDuration = gconv.Duration(defaultExtendMaxDuration)
	}
	if options.LockTimeout <= 0 {
		options.LockTimeout = gconv.Duration(defaultLockTimeout)
	}
	if options.UnlockTimeout <= 0 {
		options.UnlockTimeout = gconv.Duration(defaultUnlockTimeout)
	}

	return options
}

type Option func(*Options)

func WithExpiry(expiry time.Duration) Option {
	return func(options *Options) {
		if expiry > 0 {
			options.Expiry = expiry
		}
	}
}

func WithTries(tries int) Option {
	return func(options *Options) {
		if tries > 0 {
			options.Tries = tries
		}
	}
}

func WithRetryDelay(retryDelay time.Duration) Option {
	return func(options *Options) {
		if retryDelay > 0 {
			options.RetryDelay = retryDelay
		}
	}
}

func WithExtendMaxDuration(extendMaxDuration time.Duration) Option {
	return func(options *Options) {
		if extendMaxDuration >= 0 {
			options.ExtendMaxDuration = extendMaxDuration
		}
	}
}

func WithLockTimeout(lockTimeout time.Duration) Option {
	return func(options *Options) {
		if lockTimeout >= 0 {
			options.LockTimeout = lockTimeout
		}
	}
}

func WithUnlockTimeout(unlockTimeout time.Duration) Option {
	return func(options *Options) {
		if unlockTimeout > 0 {
			options.UnlockTimeout = unlockTimeout
		}
	}
}
