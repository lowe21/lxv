package graylog

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	defaultAddress           = "127.0.0.1:12201"
	defaultChunkSize         = 8192
	defaultWorkerNumber      = 1
	defaultReconnectInterval = "5s"
	defaultVersion           = "1.1"
)

type Options struct {
	Address           string
	ChunkSize         int
	WorkerNumber      int
	ReconnectInterval time.Duration
	Version           string
}

func defaultOptions() *Options {
	options := &Options{}
	if err := g.Config().MustGet(nil, "graylog").Scan(options); err != nil {
		panic(err)
	}

	if options.Address == "" {
		options.Address = defaultAddress
	}
	if options.ChunkSize <= 0 {
		options.ChunkSize = defaultChunkSize
	}
	if options.WorkerNumber <= 0 {
		options.WorkerNumber = defaultWorkerNumber
	}
	if options.ReconnectInterval <= 0 {
		options.ReconnectInterval = gconv.Duration(defaultReconnectInterval)
	}
	if options.Version == "" {
		options.Version = defaultVersion
	}

	return options
}
