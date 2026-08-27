package rabbitmq

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	defaultUri               = ""
	defaultVhost             = "/"
	defaultChannelMax        = 0
	defaultFrameSize         = 0
	defaultHeartbeat         = "30s"
	defaultReconnectMax      = 10
	defaultReconnectInterval = "5s"
	defaultRetryMax          = 0
	defaultRetryFactor       = 2
	defaultRetryIntervalMin  = "3s"
	defaultRetryIntervalMax  = "30s"
	defaultConsumeConcurrent = 1
	defaultConsumePrefetch   = 0
	defaultConsumeDlxSuffix  = ".dlx"
)

type Options struct {
	Product           string
	Uri               string
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
	ConsumeDlxSuffix  string
}

func defaultOptions() *Options {
	options := &Options{}
	if err := g.Config().MustGet(nil, "rabbitmq").Scan(options); err != nil {
		panic(err)
	}

	if options.Product == "" {
		options.Product = g.Server().GetName()
	}
	if options.Uri == "" {
		options.Uri = defaultUri
	}
	if options.Vhost == "" {
		options.Vhost = defaultVhost
	}
	if options.ChannelMax < 0 {
		options.ChannelMax = defaultChannelMax
	}
	if options.FrameSize < 0 {
		options.FrameSize = defaultFrameSize
	}
	if options.Heartbeat <= 0 {
		options.Heartbeat = gconv.Duration(defaultHeartbeat)
	}
	if options.ReconnectMax <= 0 {
		options.ReconnectMax = defaultReconnectMax
	}
	if options.ReconnectInterval <= 0 {
		options.ReconnectInterval = gconv.Duration(defaultReconnectInterval)
	}
	if options.RetryMax < 0 {
		options.RetryMax = defaultRetryMax
	}
	if options.RetryFactor <= 0 {
		options.RetryFactor = defaultRetryFactor
	}
	if options.RetryIntervalMin <= 0 {
		options.RetryIntervalMin = gconv.Duration(defaultRetryIntervalMin)
	}
	if options.RetryIntervalMax <= 0 {
		options.RetryIntervalMax = gconv.Duration(defaultRetryIntervalMax)
	}
	if options.ConsumeConcurrent <= 0 {
		options.ConsumeConcurrent = defaultConsumeConcurrent
	}
	if options.ConsumePrefetch < 0 {
		options.ConsumePrefetch = defaultConsumePrefetch
	}
	if options.ConsumeDlxSuffix == "" {
		options.ConsumeDlxSuffix = defaultConsumeDlxSuffix
	}

	return options
}
