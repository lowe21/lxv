package graylog

import (
	"context"
	"log"

	"github.com/gogf/gf/v2/os/grpool"
)

func New(opts ...Option) *Graylog {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	graylog := &Graylog{
		options: options,
		gelf:    make(chan *Gelf, options.WorkerNumber),
	}

	for range options.WorkerNumber {
		if err := grpool.AddWithRecover(context.Background(), func(_ context.Context) {
			graylog.worker()
		}, func(_ context.Context, exception error) {
			log.Printf("graylog worker exception %v", exception)
		}); err != nil {
			panic(err)
		}
	}

	return graylog
}
