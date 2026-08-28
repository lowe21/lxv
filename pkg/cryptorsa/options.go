package cryptorsa

import (
	"github.com/gogf/gf/v2/frame/g"
)

const (
	defaultHash       = SHA256
	defaultMinKeyBits = 2048
)

type Options struct {
	Hash       string
	MinKeyBits int
}

func defaultOptions() *Options {
	options := &Options{}
	if err := g.Config().MustGet(nil, "crypto.rsa").Scan(options); err != nil {
		panic(err)
	}

	if options.Hash == "" {
		options.Hash = defaultHash
	}
	if options.MinKeyBits <= 0 {
		options.MinKeyBits = defaultMinKeyBits
	}

	return options
}

type Option func(*Options)

func WithHash(hash string) Option {
	return func(options *Options) {
		if hash != "" {
			options.Hash = hash
		}
	}
}
