package graylog

import (
	"sync"
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
			gelf:    make(chan *Gelf, options.QueueSize),
		}

		for range options.WorkerNumber {
			go graylog.worker()
		}
	})

	return graylog
}

func Send(gelf *Gelf) {
	instance().Send(gelf)
}
