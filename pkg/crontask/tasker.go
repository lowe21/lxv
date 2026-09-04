package crontask

import (
	"context"
	"fmt"
	"sync"

	"github.com/lowe21/lxv/pkg/errcode"
)

type Tasker interface {
	Name() string
	Run(ctx context.Context) error
}

var (
	taskers map[string]Tasker
	mutex   sync.RWMutex
)

func GetTasker(name string) (tasker Tasker, err error) {
	mutex.RLock()
	defer mutex.RUnlock()

	tasker, ok := taskers[name]
	if !ok {
		err = errcode.New(fmt.Sprintf("tasker not found, name: %s", name))
	}

	return
}

func SetTasker(tasker Tasker) {
	name := tasker.Name()
	if name == "" {
		panic(fmt.Sprintf("tasker name is empty, type: %T", tasker))
	}

	mutex.Lock()
	defer mutex.Unlock()

	if taskers == nil {
		taskers = make(map[string]Tasker)
	}
	if _, ok := taskers[name]; ok {
		panic(fmt.Sprintf("tasker already exists, name: %s", name))
	}
	taskers[name] = tasker
}
