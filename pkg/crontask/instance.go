package crontask

import (
	"sync"

	"github.com/gogf/gf/v2/os/gcron"
)

var (
	cronTask *CronTask
	once     sync.Once
)

func instance() *CronTask {
	once.Do(func() {
		cronTask = &CronTask{
			options: defaultOptions(),
			cron:    gcron.New(),
		}
	})

	return cronTask
}

func Start() {
	instance().Start()
}

func Stop() {
	instance().Stop()
}
