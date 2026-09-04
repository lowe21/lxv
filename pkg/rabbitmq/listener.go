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
		ConsumeDLX(ctx context.Context, message *Message) error
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
	exchangeName := queueListener.ExchangeName()
	if exchangeName == "" {
		panic(fmt.Sprintf("queueListener exchangeName is empty, type: %T", queueListener))
	}
	routingKey := queueListener.RoutingKey()
	if routingKey == "" {
		panic(fmt.Sprintf("queueListener routingKey is empty, type: %T", queueListener))
	}

	mutex.Lock()
	defer mutex.Unlock()

	if queueListeners == nil {
		queueListeners = make(map[string]QueueListener)
	}
	name := gstr.Join([]string{exchangeName, routingKey}, ".")
	if _, ok := queueListeners[name]; ok {
		panic(fmt.Sprintf("queueListener already exists, exchangeName: %s, routingKey: %s", exchangeName, routingKey))
	}
	queueListeners[name] = queueListener
}

func SetSubscribeListener(subscribeListener SubscribeListener) {
	exchangeName := subscribeListener.ExchangeName()
	if exchangeName == "" {
		panic(fmt.Sprintf("subscribeListener exchangeName is empty, type: %T", subscribeListener))
	}

	mutex.Lock()
	defer mutex.Unlock()

	if subscribeListeners == nil {
		subscribeListeners = make(map[string]SubscribeListener)
	}
	if _, ok := subscribeListeners[exchangeName]; ok {
		panic(fmt.Sprintf("subscribeListener already exists, exchangeName: %s", exchangeName))
	}
	subscribeListeners[exchangeName] = subscribeListener
}
