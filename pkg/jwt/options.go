package jwt

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	key     = "64f17cb03b83fe8dc188865b5a250920"
	expires = "30d"
	leeway  = "7d"
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
		options.Key = []byte(key)
	}
	if options.Expires <= 0 {
		options.Expires = gconv.Duration(expires)
	}
	if options.Leeway <= 0 {
		options.Leeway = gconv.Duration(leeway)
	}

	return options
}
