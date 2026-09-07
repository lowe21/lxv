package cryptorsa

import (
	"github.com/gogf/gf/v2/frame/g"
)

const (
	hashType   = SHA256
	minKeyBits = 2048
)

type Options struct {
	HashType   string
	MinKeyBits int
}

func defaultOptions() *Options {
	options := &Options{}
	if err := g.Config().MustGet(nil, "crypto.rsa").Scan(options); err != nil {
		panic(err)
	}

	if options.HashType == "" {
		options.HashType = hashType
	}
	if options.MinKeyBits <= 0 {
		options.MinKeyBits = minKeyBits
	}

	return options
}

type Option func(*Options)

func WithHashType(hashType string) Option {
	return func(options *Options) {
		if hashType != "" {
			options.HashType = hashType
		}
	}
}
