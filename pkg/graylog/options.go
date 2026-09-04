package graylog

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	address           = "127.0.0.1:12201"
	chunkSize         = 8192
	queueSize         = 32
	workerNumber      = 1
	reconnectInterval = "5s"
	version           = "1.1"
)

type Options struct {
	Address           string
	ChunkSize         int
	QueueSize         int
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
		options.Address = address
	}
	if options.ChunkSize <= 0 {
		options.ChunkSize = chunkSize
	}
	if options.QueueSize <= 0 {
		options.QueueSize = queueSize
	}
	if options.WorkerNumber <= 0 {
		options.WorkerNumber = workerNumber
	}
	if options.ReconnectInterval <= 0 {
		options.ReconnectInterval = gconv.Duration(reconnectInterval)
	}
	if options.Version == "" {
		options.Version = version
	}

	return options
}
