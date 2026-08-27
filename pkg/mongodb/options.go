package mongodb

import (
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	defaultGroup           = "default"
	defaultUri             = ""
	defaultConnectTimeout  = "3s"
	defaultMaxConnIdleTime = "0"
	defaultMaxPoolSize     = 0
	defaultMinPoolSize     = 0
	defaultMaxConnecting   = 0
)

type Options struct {
	Uri             string
	AppName         string
	Database        string
	ConnectTimeout  time.Duration
	MaxConnIdleTime time.Duration
	MaxPoolSize     int64
	MinPoolSize     int64
	MaxConnecting   int64
}

func defaultOptions(group string) *Options {
	config := g.Config().MustGet(nil, gstr.Join([]string{"mongodb", group}, "."))
	if config.IsEmpty() {
		panic(fmt.Sprintf(`mongodb configuration node "%s" is not found, did you misspell group name "%s" or miss the mongodb configuration?`, group, group))
	}

	options := &Options{}
	if err := config.Scan(options); err != nil {
		panic(err)
	}

	if options.Uri == "" {
		options.Uri = defaultUri
	}
	if options.AppName == "" {
		options.AppName = g.Server().GetName()
	}
	if options.Database == "" {
		panic("database is empty")
	}
	if options.ConnectTimeout <= 0 {
		options.ConnectTimeout = gconv.Duration(defaultConnectTimeout)
	}
	if options.MaxConnIdleTime < 0 {
		options.MaxConnIdleTime = gconv.Duration(defaultMaxConnIdleTime)
	}
	if options.MaxPoolSize < 0 {
		options.MaxPoolSize = defaultMaxPoolSize
	}
	if options.MinPoolSize < 0 {
		options.MinPoolSize = defaultMinPoolSize
	}
	if options.MaxConnecting < 0 {
		options.MaxConnecting = defaultMaxConnecting
	}

	return options
}
