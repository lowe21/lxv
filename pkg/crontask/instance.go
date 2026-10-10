package crontask

import (
	"context"
	"sync"
)

var (
	cronTask *CronTask
	mutex    sync.Mutex
)

func instance() *CronTask {
	mutex.Lock()
	defer mutex.Unlock()

	if cronTask == nil {
		cronTask = &CronTask{
			options: newOptions(),
		}
	}

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
