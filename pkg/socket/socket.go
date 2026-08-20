package socket

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/lowe21/lxv/pkg/error_code"
)

type Socket struct {
	options   *Options
	upgrader  websocket.Upgrader
	redis     *gredis.Redis
	register  *Register
	connector *Connector
	notifier  *Notifier
	ctx       context.Context
	cancel    context.CancelFunc
	once      sync.Once
}

func (s *Socket) Start() {
	s.once.Do(func() {
		s.ctx, s.cancel = context.WithCancel(context.Background())

		if err := s.register.AddNode(s.ctx); err != nil {
			g.Log().Errorf(s.ctx, "register node error: %v", err)
		}

		go s.register.Heartbeat(s.ctx)
		go s.notifier.Subscribe(s.ctx)
	})
}

func (s *Socket) Connect(request *ghttp.Request, clientId string, group ...string) (err error) {
	if clientId == "" {
		return error_code.New(error_code.InvalidRequest, "client id is empty")
	}

	conn, err := s.upgrader.Upgrade(request.Response.Writer, request.Request, nil)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.WithoutCancel(request.GetCtx()))

	client := &Client{
		socket: s,
		conn:   conn,
		id:     clientId,
		group:  s.connector.groupName(group...),
		input:  make(chan []byte, s.options.InputQueueSize),
		output: make(chan []byte, s.options.OutputQueueSize),
		done:   make(chan struct{}),
		ctx:    ctx,
		cancel: cancel,
	}
	client.Start()

	if err = s.connector.AddClient(request.GetCtx(), client); err != nil {
		client.Send(Message(client.id, "error", "connect failed"))
		client.Close([]byte("system error"))
	} else {
		client.Send(Message(client.id, "connect", "connect succeed"))
	}

	return
}

func (s *Socket) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	for _, group := range s.connector.GetGroups() {
		for _, client := range s.connector.GetClients(group) {
			client.Close([]byte("shutdown"))
		}
	}
	_ = s.register.DeleteNode()
}
