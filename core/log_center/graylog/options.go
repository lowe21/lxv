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
)

type Options struct {
	Address           string
	ChunkSize         int
	WorkerNumber      int
	ReconnectInterval time.Duration
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

	return options
}

type Option func(*Options)

func WithAddress(address string) Option {
	return func(options *Options) {
		if address == "" {
			address = defaultAddress
		}
		options.Address = address
	}
}

func WithChunkSize(chunkSize int) Option {
	return func(options *Options) {
		if chunkSize <= 0 {
			chunkSize = defaultChunkSize
		}
		options.ChunkSize = chunkSize
	}
}

func WorkerNumber(workerNumber int) Option {
	return func(options *Options) {
		if workerNumber <= 0 {
			workerNumber = defaultWorkerNumber
		}
		options.WorkerNumber = workerNumber
	}
}

func WorkerReconnectInterval(reconnectInterval time.Duration) Option {
	return func(options *Options) {
		if reconnectInterval <= 0 {
			reconnectInterval = gconv.Duration(defaultReconnectInterval)
		}
		options.ReconnectInterval = reconnectInterval
	}
}
