package rabbitmq

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/text/gstr"
)

type RabbitMq struct {
	options    *Options
	connection *amqp.Connection
	producer   *Producer
	consumer   *Consumer
	ctx        context.Context
	cancel     context.CancelFunc
	mutex      sync.RWMutex
	once       sync.Once
}

func (r *RabbitMq) Start() {
	r.once.Do(func() {
		r.ctx, r.cancel = context.WithCancel(context.Background())

		rows := []string{"#", "EXCHANGE TYPE", "EXCHANGE NAME", "ROUTING KEY", "LISTENER", "STATUS"}
		table := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Merging: tw.CellMerging{Mode: tw.MergeHorizontal},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignCenter}},
			},
		}))
		table.Header(garray.New().Pad(len(rows), "RABBITMQ").Slice())
		if err := table.Append(rows); err != nil {
			panic(err)
		}
		defer func() {
			if len(queueListeners) > 0 || len(subscribeListeners) > 0 {
				_ = table.Render()
			} else {
				_ = table.Close()
			}
		}()

		index := 0
		for _, queueListener := range queueListeners {
			exchangeName := queueListener.ExchangeName()
			routingKey := queueListener.RoutingKey()
			if err := r.consumer.Listen(r.ctx, amqp.ExchangeDirect, exchangeName, routingKey, queueListener); err != nil {
				panic(err)
			}
			index += 1
			if err := table.Append(index, amqp.ExchangeDirect, exchangeName, routingKey, fmt.Sprintf("%T", queueListener), "LISTEN"); err != nil {
				panic(err)
			}
		}

		for _, subscribeListener := range subscribeListeners {
			exchangeName := subscribeListener.ExchangeName()
			if err := r.consumer.Listen(r.ctx, amqp.ExchangeFanout, exchangeName, "", subscribeListener); err != nil {
				panic(err)
			}
			index += 1
			if err := table.Append(index, amqp.ExchangeFanout, exchangeName, "", fmt.Sprintf("%T", subscribeListener), "LISTEN"); err != nil {
				panic(err)
			}
		}
	})
}

func (r *RabbitMq) Connection() (connection *amqp.Connection, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.connection == nil || r.connection.IsClosed() {
		properties := amqp.NewConnectionProperties()
		properties["product"] = r.options.Product

		r.connection, err = amqp.DialConfig(r.options.Uri, amqp.Config{
			Vhost:      r.options.Vhost,
			ChannelMax: uint16(r.options.ChannelMax),
			FrameSize:  r.options.FrameSize,
			Heartbeat:  r.options.Heartbeat,
			Properties: properties,
			Recovery: &amqp.Recovery{
				ReconnectionConfig: &amqp.ReconnectionConfig{
					MaxRetryCount: r.options.ReconnectMax,
					RetryInterval: r.options.ReconnectInterval,
				},
			},
		})
		if err != nil {
			return
		}
	}

	return r.connection, nil
}

func (r *RabbitMq) Channel() (channel *amqp.Channel, err error) {
	connection, err := r.Connection()
	if err != nil {
		return
	}

	return connection.Channel()
}

func (r *RabbitMq) ExchangeName(exchangeName string) (name string) {
	if r.options.Product != "" && exchangeName != "" {
		return gstr.Join([]string{r.options.Product, exchangeName}, ".")
	}

	return exchangeName
}

func (r *RabbitMq) RoutingKey(routingKey, suffix string) (key string) {
	if routingKey != "" && suffix != "" {
		return gstr.Join([]string{routingKey, suffix}, "")
	}

	return routingKey
}

func (r *RabbitMq) QueueName(exchangeName, routingKey string, suffix ...string) (name string) {
	defer func() {
		if name != "" {
			name = gstr.Join(append([]string{name}, suffix...), "")
		}
	}()

	if exchangeName != "" && routingKey != "" {
		return gstr.Join([]string{exchangeName, routingKey}, ".")
	}

	return routingKey
}

func (r *RabbitMq) QueueCreate(channel *amqp.Channel, exchangeType, exchangeName, routingKey string, exclusive bool, args amqp.Table) (queue amqp.Queue, err error) {
	exchangeName = r.ExchangeName(exchangeName)

	if err = channel.ExchangeDeclare(exchangeName, "x-delayed-message", true, false, false, false, amqp.Table{
		"x-delayed-type": exchangeType,
	}); err != nil {
		return
	}

	queue, err = channel.QueueDeclare(r.QueueName(exchangeName, routingKey), true, false, exclusive, false, args)
	if err != nil {
		return
	}
	if err = channel.QueueBind(queue.Name, routingKey, exchangeName, false, nil); err != nil {
		return
	}

	return
}

func (r *RabbitMq) QueueDelete(channel *amqp.Channel, exchangeName, routingKey string) (err error) {
	exchangeName = r.ExchangeName(exchangeName)

	if _, err = channel.QueueDelete(r.QueueName(exchangeName, routingKey), false, false, false); err != nil {
		return
	}
	if _, err = channel.QueueDelete(r.QueueName(exchangeName, routingKey, r.options.ConsumeDlxSuffix), false, false, false); err != nil {
		return
	}

	return
}

func (r *RabbitMq) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
	r.consumer.wg.Wait()

	r.mutex.RLock()
	if r.connection != nil {
		_ = r.connection.Close()
	}
	r.mutex.RUnlock()
}
