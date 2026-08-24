package crypto_rsa

import (
	"github.com/gogf/gf/v2/frame/g"
)

const (
	Sha1      = "sha1"
	Sha256    = "sha256"
	Sha384    = "sha384"
	Sha512    = "sha512"
	PssSha256 = "pss-sha256"
	PssSha384 = "pss-sha384"
	PssSha512 = "pss-sha512"
)

const (
	defaultHash       = Sha256
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
	if options.MinKeyBits <= defaultMinKeyBits {
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
