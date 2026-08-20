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

					parts, err := url.Parse(origin)
					if err != nil {
						return false
					}
					for _, host := range options.AllowedOrigins {
						if host == parts.Host {
							return true
						}
					}

					return false
				},
			},
			redis: g.Redis(options.RedisGroup),
		}
		socket.register = &Register{
			socket: socket,
		}
		socket.connector = &Connector{
			socket:  socket,
			clients: make(map[string]map[string]*Client),
		}
		socket.notifier = &Notifier{
			socket: socket,
		}
		socket.Start()
	})

	return socket
}

func Connect(request *ghttp.Request, clientId string, group ...string) (err error) {
	return instance().Connect(request, clientId, group...)
}

func Notice(ctx context.Context, message []byte, clientIds []string, group ...string) (err error) {
	return instance().notifier.Notice(ctx, message, clientIds, group...)
}

func CloseClient(ctx context.Context, message []byte, clientIds []string, group ...string) (err error) {
	return instance().notifier.CloseClient(ctx, message, "", clientIds, group...)
}

func Stop() {
	instance().Stop()
}
