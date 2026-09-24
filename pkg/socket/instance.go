package socket

import (
	"context"
	"net/http"
	"slices"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var (
	socket *Socket
	once   sync.Once
)

func instance() *Socket {
	once.Do(func() {
		options := defaultOptions()
		socket = &Socket{
			options: options,
			upgrader: websocket.Upgrader{
				CheckOrigin: func(request *http.Request) bool {
					if len(options.AllowedOrigins) > 0 {
						if slices.Contains(options.AllowedOrigins, "*") {
							return true
						}
						if origin := request.Header.Get("Origin"); origin != "" {
							return slices.Contains(options.AllowedOrigins, origin)
						}
					}
					return false
				},
			},
			redis: g.Redis(options.RedisGroup),
		}
		socket.register = &Register{
			Socket: socket,
		}
		socket.connector = &Connector{
			Socket:  socket,
			clients: make(map[string]map[string]*Client),
		}
		socket.broadcaster = &Broadcaster{
			Socket: socket,
		}
	})

	return socket
}

func Start() {
	instance().Start()
}

func Connect(request *ghttp.Request, clientID string, group ...string) error {
	return instance().Connect(request, clientID, group...)
}

func Notice(ctx context.Context, message []byte, clientIDs []string, group ...string) error {
	return instance().broadcaster.Notice(ctx, message, clientIDs, group...)
}

func CloseClient(ctx context.Context, message []byte, clientIDs []string, group ...string) error {
	return instance().broadcaster.CloseClient(ctx, message, "", clientIDs, nil, group...)
}

func Stop() {
	instance().Stop()
}
