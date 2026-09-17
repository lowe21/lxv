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
		options := newOptions()

		graylog = &Graylog{
			options: options,
			gelf:    make(chan *Gelf, options.MaxQueueSize),
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
