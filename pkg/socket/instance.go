package socket

import (
	"context"
	"net/http"
	"net/url"
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
					if len(options.AllowedOrigins) == 0 {
						return true
					}

					origin := request.Header.Get("Origin")
					if origin == "" {
						return true
					}

					if parsed, _ := url.Parse(origin); parsed != nil {
						for _, host := range options.AllowedOrigins {
							if host == parsed.Host {
								return true
							}
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

func Connect(request *ghttp.Request, clientId string, group ...string) (err error) {
	return instance().Connect(request, clientId, group...)
}

func Notice(ctx context.Context, message []byte, clientIds []string, group ...string) (err error) {
	return instance().broadcaster.Notice(ctx, message, clientIds, group...)
}

func CloseClient(ctx context.Context, message []byte, clientIds []string, group ...string) (err error) {
	return instance().broadcaster.CloseClient(ctx, message, "", clientIds, group...)
}

func Stop() {
	instance().Stop()
}
