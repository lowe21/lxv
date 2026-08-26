package rabbitmq

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/encoding/gjson"
)

var (
	rabbitMq *RabbitMq
	once     sync.Once
)

func instance() *RabbitMq {
	once.Do(func() {
		options := defaultOptions()

		rabbitMq = &RabbitMq{
			options: options,
		}
		rabbitMq.producer = &Producer{
			RabbitMq: rabbitMq,
		}
		rabbitMq.consumer = &Consumer{
			RabbitMq: rabbitMq,
		}
	})

	return rabbitMq
}

func Start() {
	instance().Start()
}

func Publish(ctx context.Context, exchangeName, routingKey string, message *Message, opts ...ProducerOption) (err error) {
	return instance().producer.Publish(ctx, exchangeName, routingKey, gjson.MustEncode(message), opts...)
}

func Broadcast(ctx context.Context, exchangeName string, message *Message, opts ...ProducerOption) (err error) {
	return instance().producer.Broadcast(ctx, exchangeName, gjson.MustEncode(message), opts...)
}

func Stop() {
	instance().Stop()
}
