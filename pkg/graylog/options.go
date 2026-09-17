package graylog

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	address           = "127.0.0.1:12201"
	workerNumber      = 1
	maxQueueSize      = 32
	maxChunkSize      = 8192
	reconnectInterval = "5s"
)

type Options struct {
	Address           string
	WorkerNumber      int
	MaxQueueSize      int
	MaxChunkSize      int
	ReconnectInterval time.Duration
}

func newOptions() *Options {
	options := &Options{}
	if err := g.Config().MustGet(nil, "graylog").Scan(options); err != nil {
		panic(err)
	}

	if options.Address == "" {
		options.Address = address
	}
	if options.WorkerNumber <= 0 {
		options.WorkerNumber = workerNumber
	}
	if options.MaxQueueSize <= 0 {
		options.MaxQueueSize = maxQueueSize
	}
	if options.MaxChunkSize <= 0 {
		options.MaxChunkSize = maxChunkSize
	}
	if options.ReconnectInterval <= 0 {
		options.ReconnectInterval = gconv.Duration(reconnectInterval)
	}

	return options
}
