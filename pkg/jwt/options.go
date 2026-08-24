package jwt

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	defaultKey     = "64f17cb03b83fe8dc188865b5a250920"
	defaultExpires = "30d"
	defaultLeeway  = "7d"
)

type Options struct {
	Issuer  string
	Key     []byte
	Expires time.Duration
	Leeway  time.Duration
}

func defaultOptions() *Options {
	options := &Options{}
	if err := g.Config().MustGet(nil, "jwt").Scan(options); err != nil {
		panic(err)
	}

	if options.Issuer == "" {
		options.Issuer = g.Server().GetName()
	}
	if len(options.Key) == 0 {
		options.Key = []byte(defaultKey)
	}
	if options.Expires <= 0 {
		options.Expires = gconv.Duration(defaultExpires)
	}
	if options.Leeway <= 0 {
		options.Leeway = gconv.Duration(defaultLeeway)
	}

	return options
}
