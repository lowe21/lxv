package rabbitmq

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gogf/gf/v2/util/gconv"

	"github.com/lowe21/lxv/pkg/errcode"
)

type Producer struct {
	*RabbitMQ
	connection *amqp.Connection
	mutex      sync.RWMutex
}

func (p *Producer) Channel() (channel *amqp.Channel, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.connection == nil || p.connection.IsClosed() {
		p.connection, err = p.Connection()
		if err != nil {
			return
		}
	}

	return p.connection.Channel()
}

func (p *Producer) Publish(ctx context.Context, exchangeName, routingKey string, body []byte, opts ...ProducerOption) (err error) {
	channel, err := p.Channel()
	if err != nil {
		return
	}
	defer func() {
		_ = channel.Close()
	}()

	if err = channel.Confirm(false); err != nil {
		return
	}

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

	confirmation, err := channel.PublishWithDeferredConfirmWithContext(ctx, p.ExchangeName(exchangeName), routingKey, false, false, amqp.Publishing{
		Headers: amqp.Table{
			"x-delay":       options.Delay,
			"x-retry-count": options.RetryCount,
		},
		DeliveryMode: amqp.Persistent,
		Expiration:   expiration,
		Body:         body,
	})
	if err != nil {
		return
	}

	acked, err := confirmation.WaitContext(ctx)
	if err != nil {
		return
	}
	if !acked {
		err = errcode.New("publish is not confirmed")
	}

	return
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
	Expiration int64
	Delay      int64
	RetryCount int
}

type ProducerOption func(*ProducerOptions)

func WithExpiration(expiration int64) ProducerOption {
	return func(options *ProducerOptions) {
		if expiration >= 0 {
			options.Expiration = expiration
		}
	}
}

func WithDelay(delay int64) ProducerOption {
	return func(options *ProducerOptions) {
		if delay > 0 {
			options.Delay = delay
		}
	}
}

func withRetryCount(retryCount int) ProducerOption {
	return func(options *ProducerOptions) {
		if retryCount > 0 {
			options.RetryCount = retryCount
		}
	}
}
