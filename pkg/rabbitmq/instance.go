package rabbitmq

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/encoding/gjson"
)

var (
	rabbitMQ *RabbitMQ
	mutex    sync.Mutex
)

func instance() *RabbitMQ {
	mutex.Lock()
	defer mutex.Unlock()

	if rabbitMQ == nil {
		options := newOptions()
		rabbitMQ = &RabbitMQ{
			options: options,
		}
		rabbitMQ.producer = &Producer{
			RabbitMQ: rabbitMQ,
		}
		rabbitMQ.consumer = &Consumer{
			RabbitMQ: rabbitMQ,
		}
	}

	return rabbitMQ
}

func Start() {
	instance().Start()
}

func Publish(ctx context.Context, exchangeName, routingKey string, message *Message, opts ...ProducerOption) error {
	return instance().producer.Publish(ctx, exchangeName, routingKey, gjson.MustEncode(message), opts...)
}

func Broadcast(ctx context.Context, exchangeName string, message *Message, opts ...ProducerOption) error {
	return instance().producer.Broadcast(ctx, exchangeName, gjson.MustEncode(message), opts...)
}

func Stop() {
	instance().Stop()
}
