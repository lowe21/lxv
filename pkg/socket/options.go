package socket

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
)

const (
	defaultRedisGroup         = "default"
	defaultRedisKeyPrefix     = "socket"
	defaultRedisChannel       = "broadcast"
	defaultNodeTtl            = "60s"
	defaultNodeHeartbeat      = "20s"
	defaultClientDefaultGroup = "default"
	defaultInputQueueSize     = 64
	defaultOutputQueueSize    = 64
	defaultMessageMaxSize     = 512
	defaultPingInterval       = "60s"
	defaultPongTimeout        = "90s"
	defaultWriteTimeout       = "10s"
)

type Options struct {
	RedisGroup         string
	RedisKeyPrefix     string
	RedisChannel       string
	NodeId             string
	NodeTtl            time.Duration
	NodeHeartbeat      time.Duration
	AllowedOrigins     []string
	ClientDefaultGroup string
	InputQueueSize     int
	OutputQueueSize    int
	MessageMaxSize     int64
	PingInterval       time.Duration
	PongTimeout        time.Duration
	WriteTimeout       time.Duration
}

func defaultOptions() *Options {
	options := &Options{}
	if err := g.Config().MustGet(nil, "socket").Scan(options); err != nil {
		panic(err)
	}

	if options.RedisGroup == "" {
		options.RedisGroup = defaultRedisGroup
	}
	if options.RedisKeyPrefix == "" {
		options.RedisKeyPrefix = defaultRedisKeyPrefix
	}
	if options.RedisChannel == "" {
		options.RedisChannel = defaultRedisChannel
	}
	if options.NodeId == "" {
		options.NodeId = guid.S()
	}
	if options.NodeTtl <= 0 {
		options.NodeTtl = gconv.Duration(defaultNodeTtl)
	}
	if options.NodeHeartbeat <= 0 {
		options.NodeHeartbeat = gconv.Duration(defaultNodeHeartbeat)
	}
	if options.NodeHeartbeat >= options.NodeTtl {
		panic("nodeHeartbeat must be less than nodeTtl")
	}
	if options.ClientDefaultGroup == "" {
		options.ClientDefaultGroup = defaultClientDefaultGroup
	}
	if options.InputQueueSize <= 0 {
		options.InputQueueSize = defaultInputQueueSize
	}
	if options.OutputQueueSize <= 0 {
		options.OutputQueueSize = defaultOutputQueueSize
	}
	if options.MessageMaxSize <= 0 {
		options.MessageMaxSize = defaultMessageMaxSize
	}
	if options.PingInterval <= 0 {
		options.PingInterval = gconv.Duration(defaultPingInterval)
	}
	if options.PongTimeout <= 0 {
		options.PongTimeout = gconv.Duration(defaultPongTimeout)
	}
	if options.PongTimeout < options.PingInterval {
		panic("pongTimeout must be greater than or equal to pingInterval")
	}
	if options.WriteTimeout <= 0 {
		options.WriteTimeout = gconv.Duration(defaultWriteTimeout)
	}

	return options
}
