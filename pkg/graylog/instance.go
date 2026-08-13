package graylog

import (
	"context"
	"log"
	"sync"

	"github.com/gogf/gf/v2/os/grpool"
)

var (
	graylog *Graylog
	once    sync.Once
)

func instance() *Graylog {
	once.Do(func() {
		options := defaultOptions()

		graylog = &Graylog{
			options: options,
			gelf:    make(chan *Gelf, options.WorkerNumber),
		}

		for range options.WorkerNumber {
			if err := grpool.AddWithRecover(context.Background(), func(_ context.Context) {
				graylog.worker()
			}, func(_ context.Context, exception error) {
				log.Printf("graylog worker exception: %v", exception)
			}); err != nil {
				panic(err)
			}
		}
	})

	return graylog
}

func Send(ctx context.Context, gelf *Gelf) {
	instance().Send(ctx, gelf)
}
