package rabbitmq

import (
	"context"
	"fmt"
	"net"
	"os"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gogf/gf/v2/text/gstr"
)

type RabbitMQ struct {
	options  *Options
	producer *Producer
	consumer *Consumer
	ctx      context.Context
	cancel   context.CancelFunc
	mutex    sync.RWMutex
	started  atomic.Bool
}

func (r *RabbitMQ) Start() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.started.Load() {
		rows := []string{"#", "EXCHANGE TYPE", "EXCHANGE NAME", "ROUTING KEY", "LISTENER", "STATUS"}
		table := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Merging: tw.CellMerging{Mode: tw.MergeBoth},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignCenter}},
			},
		}))
		table.Header(slices.Repeat([]string{"RABBITMQ"}, len(rows)))
		if err := table.Append(rows); err != nil {
			panic(err)
		}

		r.ctx, r.cancel = context.WithCancel(context.Background())
		r.started.Store(true)

		index := 0
		for _, queueListener := range getQueueListeners() {
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

		for _, subscribeListener := range getSubscribeListeners() {
			exchangeName := subscribeListener.ExchangeName()
			if err := r.consumer.Listen(r.ctx, amqp.ExchangeFanout, exchangeName, "", subscribeListener); err != nil {
				panic(err)
			}
			index += 1
			if err := table.Append(index, amqp.ExchangeFanout, exchangeName, "", fmt.Sprintf("%T", subscribeListener), "LISTEN"); err != nil {
				panic(err)
			}
		}

		if index > 0 {
			_ = table.Render()
		} else {
			_ = table.Close()
		}
	}
}

func (r *RabbitMQ) Connection() (connection *amqp.Connection, err error) {
	properties := amqp.NewConnectionProperties()
	properties["product"] = r.options.Product

	return amqp.DialConfig(r.options.URI, amqp.Config{
		Vhost:      r.options.Vhost,
		ChannelMax: uint16(r.options.ChannelMax),
		FrameSize:  r.options.FrameSize,
		Heartbeat:  r.options.Heartbeat,
		Properties: properties,
		Dial: func(network, address string) (net.Conn, error) {
			return (&net.Dialer{
				Timeout: r.options.DialTimeout,
			}).Dial(network, address)
		},
	})
}

func (r *RabbitMQ) ExchangeName(exchangeName string) (name string) {
	if r.options.Product != "" && exchangeName != "" {
		name = r.options.Product + "." + exchangeName
	} else {
		name = exchangeName
	}

	return
}

func (r *RabbitMQ) RoutingKey(routingKey, suffix string) (key string) {
	if routingKey != "" && suffix != "" {
		key = routingKey + suffix
	} else {
		key = routingKey
	}

	return
}

func (r *RabbitMQ) QueueName(exchangeName, routingKey string, suffix ...string) (name string) {
	if exchangeName != "" && routingKey != "" {
		name = exchangeName + "." + routingKey
	} else {
		name = routingKey
	}

	if name != "" && len(suffix) > 0 {
		name = gstr.Join(append([]string{name}, suffix...), "")
	}

	return
}

func (r *RabbitMQ) QueueCreate(channel *amqp.Channel, exchangeType, exchangeName, routingKey string, exclusive bool, args amqp.Table) (queue amqp.Queue, err error) {
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

	err = channel.QueueBind(queue.Name, routingKey, exchangeName, false, nil)

	return
}

func (r *RabbitMQ) QueueDelete(channel *amqp.Channel, exchangeName, routingKey string) (err error) {
	exchangeName = r.ExchangeName(exchangeName)

	if _, err = channel.QueueDelete(r.QueueName(exchangeName, routingKey), false, false, false); err != nil {
		return
	}

	_, err = channel.QueueDelete(r.QueueName(exchangeName, routingKey, r.options.ConsumeDLXSuffix), false, false, false)

	return
}

func (r *RabbitMQ) Stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.started.Store(false)

	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}

	r.consumer.mutex.Lock()
	if r.consumer.connection != nil {
		_ = r.consumer.connection.Close()
		r.consumer.connection = nil
	}
	r.consumer.mutex.Unlock()

	r.consumer.wg.Wait()
}
