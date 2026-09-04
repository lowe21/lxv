package socket

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
)

const (
	redisGroup         = "default"
	redisKeyPrefix     = "socket"
	redisChannel       = "broadcast"
	nodeTTL            = "60s"
	nodeHeartbeat      = "20s"
	defaultClientGroup = "default"
	inputQueueSize     = 64
	outputQueueSize    = 64
	messageMaxSize     = 512
	pingInterval       = "60s"
	pongTimeout        = "90s"
	writeTimeout       = "10s"
)

type Options struct {
	RedisGroup         string
	RedisKeyPrefix     string
	RedisChannel       string
	NodeID             string
	NodeTTL            time.Duration
	NodeHeartbeat      time.Duration
	AllowedOrigins     []string
	DefaultClientGroup string
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
		options.RedisGroup = redisGroup
	}
	if options.RedisKeyPrefix == "" {
		options.RedisKeyPrefix = redisKeyPrefix
	}
	if options.RedisChannel == "" {
		options.RedisChannel = redisChannel
	}
	if options.NodeID == "" {
		options.NodeID = guid.S()
	}
	if options.NodeTTL <= 0 {
		options.NodeTTL = gconv.Duration(nodeTTL)
	}
	if options.NodeHeartbeat <= 0 {
		options.NodeHeartbeat = gconv.Duration(nodeHeartbeat)
	}
	if options.NodeHeartbeat >= options.NodeTTL {
		panic("options error, nodeHeartbeat must be less than nodeTTL")
	}
	if options.DefaultClientGroup == "" {
		options.DefaultClientGroup = defaultClientGroup
	}
	if options.InputQueueSize <= 0 {
		options.InputQueueSize = inputQueueSize
	}
	if options.OutputQueueSize <= 0 {
		options.OutputQueueSize = outputQueueSize
	}
	if options.MessageMaxSize <= 0 {
		options.MessageMaxSize = messageMaxSize
	}
	if options.PingInterval <= 0 {
		options.PingInterval = gconv.Duration(pingInterval)
	}
	if options.PongTimeout <= 0 {
		options.PongTimeout = gconv.Duration(pongTimeout)
	}
	if options.PongTimeout < options.PingInterval {
		panic("options error, pongTimeout must be greater than or equal to pingInterval")
	}
	if options.WriteTimeout <= 0 {
		options.WriteTimeout = gconv.Duration(writeTimeout)
	}

	return options
}
