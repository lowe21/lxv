package rabbitmq

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	vhost             = "/"
	channelMax        = 0
	frameSize         = 0
	heartbeat         = "30s"
	reconnectMax      = 10
	reconnectInterval = "5s"
	retryMax          = 0
	retryFactor       = 2
	retryIntervalMin  = "3s"
	retryIntervalMax  = "30s"
	consumeConcurrent = 1
	consumePrefetch   = 0
	consumeDLXSuffix  = ".dlx"
)

type Options struct {
	Product           string
	URI               string
	Vhost             string
	ChannelMax        int
	FrameSize         int
	Heartbeat         time.Duration
	ReconnectMax      int
	ReconnectInterval time.Duration
	RetryMax          int
	RetryFactor       float64
	RetryIntervalMin  time.Duration
	RetryIntervalMax  time.Duration
	ConsumeConcurrent int
	ConsumePrefetch   int
	ConsumeDLXSuffix  string
}

func defaultOptions() *Options {
	options := &Options{}
	if err := g.Config().MustGet(nil, "rabbitmq").Scan(options); err != nil {
		panic(err)
	}

	if options.Product == "" {
		options.Product = g.Server().GetName()
	}
	if options.URI == "" {
		panic("options error, uri is empty")
	}
	if options.Vhost == "" {
		options.Vhost = vhost
	}
	if options.ChannelMax < 0 {
		options.ChannelMax = channelMax
	}
	if options.FrameSize < 0 {
		options.FrameSize = frameSize
	}
	if options.Heartbeat <= 0 {
		options.Heartbeat = gconv.Duration(heartbeat)
	}
	if options.ReconnectMax <= 0 {
		options.ReconnectMax = reconnectMax
	}
	if options.ReconnectInterval <= 0 {
		options.ReconnectInterval = gconv.Duration(reconnectInterval)
	}
	if options.RetryMax < 0 {
		options.RetryMax = retryMax
	}
	if options.RetryFactor <= 0 {
		options.RetryFactor = retryFactor
	}
	if options.RetryIntervalMin <= 0 {
		options.RetryIntervalMin = gconv.Duration(retryIntervalMin)
	}
	if options.RetryIntervalMax <= 0 {
		options.RetryIntervalMax = gconv.Duration(retryIntervalMax)
	}
	if options.ConsumeConcurrent <= 0 {
		options.ConsumeConcurrent = consumeConcurrent
	}
	if options.ConsumePrefetch < 0 {
		options.ConsumePrefetch = consumePrefetch
	}
	if options.ConsumeDLXSuffix == "" {
		options.ConsumeDLXSuffix = consumeDLXSuffix
	}

	return options
}
