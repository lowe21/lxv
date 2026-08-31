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
		err = errcode.New(fmt.Sprintf(`tasker name "%s" not found`, name))
	}

	return
}

func SetTasker(tasker Tasker) {
	mutex.Lock()
	defer mutex.Unlock()

	name := tasker.Name()
	if name == "" {
		panic(fmt.Sprintf(`tasker "%T" name is empty`, tasker))
	}

	if taskers == nil {
		taskers = make(map[string]Tasker)
	}
	if _, ok := taskers[name]; ok {
		panic(fmt.Sprintf(`tasker name "%s" already exists`, name))
	}
	taskers[name] = tasker
}
