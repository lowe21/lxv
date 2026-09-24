package rabbitmq

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/lowe21/lxv/pkg/errcode"
)

type DeliveryHandler func(ctx context.Context, delivery *amqp.Delivery, isDLX bool) error

type Consumer struct {
	*RabbitMQ
	connection *amqp.Connection
	mutex      sync.RWMutex
	wg         sync.WaitGroup
}

func (c *Consumer) Channel() (channel *amqp.Channel, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.started.Load() {
		err = errcode.New("rabbitmq is not started")
		return
	}

	if c.connection == nil || c.connection.IsClosed() {
		c.connection, err = c.Connection()
		if err != nil {
			return
		}
	}

	return c.connection.Channel()
}

func (c *Consumer) Listen(ctx context.Context, exchangeType, exchangeName, routingKey string, listener any, opts ...ConsumerOption) (err error) {
	deliveryHandler := func(ctx context.Context, delivery *amqp.Delivery, isDLX bool) (err error) {
		defer func() {
			if exception := recover(); exception != nil {
				g.Log().Errorf(ctx, "deliveryHandler panic, %+v", exception)
				err = errcode.New(exception)
			} else {
				if err != nil {
					g.Log().Error(ctx, gerror.Wrap(err, fmt.Sprintf("%s %s %s", delivery.Exchange, delivery.RoutingKey, delivery.Body)))
				}
			}
		}()

		message := &Message{}
		if err = gconv.Scan(delivery.Body, message); err != nil {
			return
		}

		switch exchangeType {
		case amqp.ExchangeDirect:
			if isDLX {
				err = listener.(QueueListener).ConsumeDLX(ctx, message)
			} else {
				err = listener.(QueueListener).Consume(ctx, message)
			}
		case amqp.ExchangeFanout:
			err = listener.(SubscribeListener).Consume(ctx, message)
		}

		return
	}

	switch exchangeType {
	case amqp.ExchangeDirect:
		if err = c.Consume(ctx, exchangeName, routingKey, deliveryHandler, opts...); err != nil {
			return
		}
		err = c.ConsumeDLX(ctx, exchangeName, routingKey, deliveryHandler, opts...)
	case amqp.ExchangeFanout:
		err = c.Subscribe(ctx, exchangeName, deliveryHandler)
	}

	return
}

func (c *Consumer) Consume(ctx context.Context, exchangeName, routingKey string, deliveryHandler DeliveryHandler, opts ...ConsumerOption) (err error) {
	if deliveryHandler == nil {
		return errcode.New("deliveryHandler is nil")
	}

	options := &ConsumerOptions{
		RetryMax:          c.options.RetryMax,
		RetryFactor:       c.options.RetryFactor,
		RetryIntervalMin:  c.options.RetryIntervalMin,
		RetryIntervalMax:  c.options.RetryIntervalMax,
		ConsumeConcurrent: c.options.ConsumeConcurrent,
		ConsumePrefetch:   c.options.ConsumePrefetch,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	consume := func() (err error) {
		if err = ctx.Err(); err != nil {
			return
		}

		channel, err := c.Channel()
		if err != nil {
			return
		}
		defer func() {
			_ = channel.Close()
		}()

		queue, err := c.QueueCreate(channel, amqp.ExchangeDirect, exchangeName, routingKey, false, amqp.Table{
			"x-dead-letter-exchange":    c.ExchangeName(exchangeName),
			"x-dead-letter-routing-key": c.RoutingKey(routingKey, c.options.ConsumeDLXSuffix),
		})
		if err != nil {
			return
		}

		if err = channel.Qos(options.ConsumePrefetch, 0, false); err != nil {
			return
		}

		deliveries, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
		if err != nil {
			return
		}

		notifyCancel := channel.NotifyCancel(make(chan string, 1))
		notifyClose := channel.NotifyClose(make(chan *amqp.Error, 1))

		for {
			select {
			case delivery, ok := <-deliveries:
				if !ok {
					return &amqp.Error{
						Code:   amqp.ChannelError,
						Reason: "delivery channel is closed",
					}
				}
				func() {
					if ctx.Err() != nil {
						_ = delivery.Nack(false, true)
					} else if deliveryHandler(ctx, &delivery, false) != nil {
						delay := time.Duration(0)
						retryCount := gconv.Int(delivery.Headers["x-retry-count"])
						if retryCount >= options.RetryMax {
							delivery.RoutingKey = c.RoutingKey(delivery.RoutingKey, c.options.ConsumeDLXSuffix)
							delivery.Expiration = ""
						} else {
							delay = min(
								time.Duration(float64(options.RetryIntervalMin)*math.Pow(options.RetryFactor, float64(retryCount))),
								options.RetryIntervalMax,
							)
						}

						if c.producer.Publish(ctx, exchangeName, delivery.RoutingKey, delivery.Body, WithDelay(delay.Milliseconds()), withRetryCount(retryCount+1)) != nil {
							_ = delivery.Nack(false, true)
						} else {
							_ = delivery.Ack(false)
						}
					} else {
						_ = delivery.Ack(false)
					}
				}()
			case <-notifyCancel:
				if !options.ConsumeCancel {
					err = &amqp.Error{
						Code:   amqp.ChannelError,
						Reason: "queue is deleted or moved to another node",
					}
				}
				return
			case err = <-notifyClose:
				return
			case <-ctx.Done():
				return
			}
		}
	}

	for range options.ConsumeConcurrent {
		c.wg.Go(func() {
			for {
				if err := consume(); err != nil {
					g.Log().Errorf(ctx, "consume error, %+v", err)
					select {
					case <-time.After(c.options.ReconnectInterval):
					case <-ctx.Done():
						return
					}
				} else {
					break
				}
			}
		})
	}

	return
}

func (c *Consumer) ConsumeDLX(ctx context.Context, exchangeName, routingKey string, deliveryHandler DeliveryHandler, opts ...ConsumerOption) (err error) {
	if deliveryHandler == nil {
		return
	}

	options := &ConsumerOptions{
		ConsumePrefetch: c.options.ConsumePrefetch,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	consumeDLX := func() (err error) {
		if err = ctx.Err(); err != nil {
			return
		}

		channel, err := c.Channel()
		if err != nil {
			return
		}
		defer func() {
			_ = channel.Close()
		}()

		queue, err := c.QueueCreate(channel, amqp.ExchangeDirect, exchangeName, c.RoutingKey(routingKey, c.options.ConsumeDLXSuffix), false, nil)
		if err != nil {
			return
		}

		if err = channel.Qos(options.ConsumePrefetch, 0, false); err != nil {
			return
		}

		deliveries, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
		if err != nil {
			return
		}

		notifyCancel := channel.NotifyCancel(make(chan string, 1))
		notifyClose := channel.NotifyClose(make(chan *amqp.Error, 1))

		for {
			select {
			case delivery, ok := <-deliveries:
				if !ok {
					return &amqp.Error{
						Code:   amqp.ChannelError,
						Reason: "delivery channel is closed",
					}
				}
				func() {
					if ctx.Err() != nil {
						_ = delivery.Nack(false, true)
					} else if deliveryHandler(ctx, &delivery, true) == nil {
						_ = delivery.Ack(false)
					}
				}()
			case <-notifyCancel:
				if !options.ConsumeCancel {
					err = &amqp.Error{
						Code:   amqp.ChannelError,
						Reason: "queue is deleted or moved to another node",
					}
				}
				return
			case err = <-notifyClose:
				return
			case <-ctx.Done():
				return
			}
		}
	}

	c.wg.Go(func() {
		for {
			if err := consumeDLX(); err != nil {
				g.Log().Errorf(ctx, "consumeDLX error, %+v", err)
				select {
				case <-time.After(c.options.ReconnectInterval):
				case <-ctx.Done():
					return
				}
			} else {
				break
			}
		}
	})

	return
}

func (c *Consumer) Subscribe(ctx context.Context, exchangeName string, deliveryHandler DeliveryHandler) (err error) {
	if deliveryHandler == nil {
		return errcode.New("deliveryHandler is nil")
	}

	subscribe := func() (err error) {
		if err = ctx.Err(); err != nil {
			return
		}

		channel, err := c.Channel()
		if err != nil {
			return
		}
		defer func() {
			_ = channel.Close()
		}()

		queue, err := c.QueueCreate(channel, amqp.ExchangeFanout, exchangeName, "", true, nil)
		if err != nil {
			return
		}

		deliveries, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
		if err != nil {
			return
		}

		notifyCancel := channel.NotifyCancel(make(chan string, 1))
		notifyClose := channel.NotifyClose(make(chan *amqp.Error, 1))

		for {
			select {
			case delivery, ok := <-deliveries:
				if !ok {
					return &amqp.Error{
						Code:   amqp.ChannelError,
						Reason: "delivery channel is closed",
					}
				}
				func() {
					if ctx.Err() != nil {
						_ = delivery.Nack(false, true)
					} else if deliveryHandler(ctx, &delivery, false) != nil {
						_ = delivery.Reject(false)
					} else {
						_ = delivery.Ack(false)
					}
				}()
			case <-notifyCancel:
				return &amqp.Error{
					Code:   amqp.ChannelError,
					Reason: "queue is deleted or moved to another node",
				}
			case err = <-notifyClose:
				return
			case <-ctx.Done():
				return
			}
		}
	}

	c.wg.Go(func() {
		for {
			if err := subscribe(); err != nil {
				g.Log().Errorf(ctx, "subscribe error, %+v", err)
				select {
				case <-time.After(c.options.ReconnectInterval):
				case <-ctx.Done():
					return
				}
			} else {
				break
			}
		}
	})

	return
}

type ConsumerOptions struct {
	RetryMax          int
	RetryFactor       float64
	RetryIntervalMin  time.Duration
	RetryIntervalMax  time.Duration
	ConsumeConcurrent int
	ConsumePrefetch   int
	ConsumeCancel     bool
}

type ConsumerOption func(*ConsumerOptions)

func WithRetryMax(retryMax int) ConsumerOption {
	return func(options *ConsumerOptions) {
		if retryMax >= 0 {
			options.RetryMax = retryMax
		}
	}
}

func WithRetryFactor(retryFactor float64) ConsumerOption {
	return func(options *ConsumerOptions) {
		if retryFactor > 0 {
			options.RetryFactor = retryFactor
		}
	}
}

func WithRetryIntervalMin(retryIntervalMin time.Duration) ConsumerOption {
	return func(options *ConsumerOptions) {
		if retryIntervalMin > 0 {
			options.RetryIntervalMin = retryIntervalMin
		}
	}
}

func WithRetryIntervalMax(retryIntervalMax time.Duration) ConsumerOption {
	return func(options *ConsumerOptions) {
		if retryIntervalMax > 0 {
			options.RetryIntervalMax = retryIntervalMax
		}
	}
}

func WithConsumeConcurrent(consumeConcurrent int) ConsumerOption {
	return func(options *ConsumerOptions) {
		if consumeConcurrent > 0 {
			options.ConsumeConcurrent = consumeConcurrent
		}
	}
}

func WithConsumePrefetch(consumePrefetch int) ConsumerOption {
	return func(options *ConsumerOptions) {
		if consumePrefetch >= 0 {
			options.ConsumePrefetch = consumePrefetch
		}
	}
}

func WithConsumeCancel(consumeCancel bool) ConsumerOption {
	return func(options *ConsumerOptions) {
		options.ConsumeCancel = consumeCancel
	}
}
