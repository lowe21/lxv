package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gogf/gf/v2/util/gconv"
)

type Producer struct {
	*RabbitMQ
}

func (p *Producer) Publish(ctx context.Context, exchangeName, routingKey string, body []byte, opts ...ProducerOption) (err error) {
	channel, err := p.Channel()
	if err != nil {
		return
	}
	defer func() {
		_ = channel.Close()
	}()

	options := &ProducerOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	expiration := ""
	if options.Expiration > 0 {
		expiration = gconv.String(options.Expiration)
	}

	return channel.PublishWithContext(ctx, p.ExchangeName(exchangeName), routingKey, false, false, amqp.Publishing{
		Headers: amqp.Table{
			"x-delay":       options.Delay,
			"x-retry-count": options.RetryCount,
		},
		DeliveryMode: amqp.Persistent,
		Expiration:   expiration,
		Body:         body,
	})
}

func (p *Producer) Broadcast(ctx context.Context, exchangeName string, body []byte, opts ...ProducerOption) (err error) {
	channel, err := p.Channel()
	if err != nil {
		return
	}
	defer func() {
		_ = channel.Close()
	}()

	options := &ProducerOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	return channel.PublishWithContext(ctx, p.ExchangeName(exchangeName), "", false, false, amqp.Publishing{
		Headers: amqp.Table{
			"x-delay": options.Delay,
		},
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

type ProducerOptions struct {
	Delay      int64
	RetryCount int
	Expiration int64
}

type ProducerOption func(*ProducerOptions)

func WithDelay(delay int64) ProducerOption {
	return func(options *ProducerOptions) {
		if delay > 0 {
			options.Delay = delay
		}
	}
}

func WithRetryCount(retryCount int) ProducerOption {
	return func(options *ProducerOptions) {
		if retryCount > 0 {
			options.RetryCount = retryCount
		}
	}
}

func WithExpiration(expiration int64) ProducerOption {
	return func(options *ProducerOptions) {
		if expiration >= 0 {
			options.Expiration = expiration
		}
	}
}
