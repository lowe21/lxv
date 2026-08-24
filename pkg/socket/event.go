package socket

import (
	"context"
	"fmt"
	"sync"

	"github.com/lowe21/lxv/pkg/error_code"
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
		err = error_code.New(error_code.InvalidEvent, fmt.Sprintf(`event "%s" not found`, name))
	}

	return
}

func SetEvent(event Event) {
	mutex.Lock()
	defer mutex.Unlock()

	name := event.Name()
	if name == "" {
		panic(fmt.Sprintf("%T missing event name", event))
	}

	if events == nil {
		events = make(map[string]Event)
	}
	if _, ok := events[name]; ok {
		panic(fmt.Sprintf(`%T event name "%s" already exists`, event, name))
	}
	events[name] = event
}
