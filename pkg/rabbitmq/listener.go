package rabbitmq

import (
	"context"
	"fmt"
	"sync"

	"github.com/gogf/gf/v2/text/gstr"
)

type (
	QueueListener interface {
		ExchangeName() string
		RoutingKey() string
		Consume(ctx context.Context, message *Message) error
		ConsumeDlx(ctx context.Context, message *Message) error
	}

	SubscribeListener interface {
		ExchangeName() string
		Consume(ctx context.Context, message *Message) error
	}
)

var (
	queueListeners     map[string]QueueListener
	subscribeListeners map[string]SubscribeListener
	mutex              sync.RWMutex
)

func SetQueueListener(queueListener QueueListener) {
	mutex.Lock()
	defer mutex.Unlock()

	exchangeName := queueListener.ExchangeName()
	if exchangeName == "" {
		panic(fmt.Sprintf("%T exchangeName is empty", queueListener))
	}
	routingKey := queueListener.RoutingKey()
	if routingKey == "" {
		panic(fmt.Sprintf("%T routingKey is empty", queueListener))
	}

	if queueListeners == nil {
		queueListeners = make(map[string]QueueListener)
	}
	name := gstr.Join([]string{exchangeName, routingKey}, ".")
	if _, ok := queueListeners[name]; ok {
		panic(fmt.Sprintf(`%T exchangeName "%s" and routingKey "%s" already exists`, queueListener, exchangeName, routingKey))
	}
	queueListeners[name] = queueListener
}

func SetSubscribeListener(subscribeListener SubscribeListener) {
	mutex.Lock()
	defer mutex.Unlock()

	exchangeName := subscribeListener.ExchangeName()
	if exchangeName == "" {
		panic(fmt.Sprintf("%T exchangeName is empty", subscribeListener))
	}

	if subscribeListeners == nil {
		subscribeListeners = make(map[string]SubscribeListener)
	}
	if _, ok := subscribeListeners[exchangeName]; ok {
		panic(fmt.Sprintf(`%T exchangeName "%s" already exists`, subscribeListener, exchangeName))
	}
	subscribeListeners[exchangeName] = subscribeListener
}
