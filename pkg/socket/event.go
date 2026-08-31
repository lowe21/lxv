package socket

import (
	"context"
	"fmt"
	"sync"

	"github.com/lowe21/lxv/pkg/errcode"
)

type Event interface {
	Name() string
	Handler(ctx context.Context, input *Input) (any, error)
}

var (
	events map[string]Event
	mutex  sync.RWMutex
)

func GetEvent(name string) (event Event, err error) {
	mutex.RLock()
	defer mutex.RUnlock()

	event, ok := events[name]
	if !ok {
		err = errcode.New(errcode.ErrInvalidEvent, fmt.Sprintf("event not found, name: %s", name))
	}

	return
}

func SetEvent(event Event) {
	mutex.Lock()
	defer mutex.Unlock()

	name := event.Name()
	if name == "" {
		panic(fmt.Sprintf("event name is empty, type: %T", event))
	}

	if events == nil {
		events = make(map[string]Event)
	}
	if _, ok := events[name]; ok {
		panic(fmt.Sprintf("event already exists, name: %s", name))
	}
	events[name] = event
}
