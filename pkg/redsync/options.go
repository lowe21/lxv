package redsync

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	defaultExpiry         = "10s"
	defaultTries          = 60
	defaultRetryDelay     = "1s"
	defaultExtendDuration = "0"
)

type Options struct {
	Expiry         time.Duration
	Tries          int
	RetryDelay     time.Duration
	ExtendDuration time.Duration
}

func defaultOptions() *Options {
	options := &Options{}
	if err := g.Config().MustGet(nil, "redsync").Scan(options); err != nil {
		panic(err)
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
	if options.ExtendDuration < 0 {
		options.ExtendDuration = gconv.Duration(defaultExtendDuration)
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

func WithExtendDuration(extendDuration time.Duration) Option {
	return func(options *Options) {
		if extendDuration >= 0 {
			options.ExtendDuration = extendDuration
		}
	}
}
