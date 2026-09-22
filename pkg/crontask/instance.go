package crontask

import (
	"context"
	"sync"
)

var (
	cronTask *CronTask
	once     sync.Once
)

func instance() *CronTask {
	once.Do(func() {
		cronTask = &CronTask{
			options: newOptions(),
		}
	})

	return cronTask
}

func Start() {
	instance().Start()
}

func AddTask(ctx context.Context, name, pattern string, tasker Tasker) error {
	return instance().AddTask(ctx, name, pattern, tasker)
}

func RemoveTask(name string) error {
	return instance().RemoveTask(name)
}

func Stop() {
	instance().Stop()
}
