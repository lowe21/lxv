package jwt

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	expires = "30d"
	leeway  = "7d"
)

type Options struct {
	Key     []byte
	Issuer  string
	Expires time.Duration
	Leeway  time.Duration
}

func newOptions() *Options {
	options := &Options{}
	if err := g.Config().MustGet(nil, "jwt").Scan(options); err != nil {
		panic(err)
	}

	if len(options.Key) < 32 {
		panic("jwt key must be at least 32 bytes")
	}
	if options.Issuer == "" {
		options.Issuer = g.Server().GetName()
	}
	if options.Expires <= 0 {
		options.Expires = gconv.Duration(expires)
	}
	if options.Leeway <= 0 {
		options.Leeway = gconv.Duration(leeway)
	}

	return options
}
