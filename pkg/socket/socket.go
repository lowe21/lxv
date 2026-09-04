package socket

import (
	"context"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/lowe21/lxv/pkg/errcode"
)

type Socket struct {
	options     *Options
	upgrader    websocket.Upgrader
	redis       *gredis.Redis
	register    *Register
	connector   *Connector
	broadcaster *Broadcaster
	ctx         context.Context
	cancel      context.CancelFunc
	once        sync.Once
}

func (s *Socket) Start() {
	s.once.Do(func() {
		s.ctx, s.cancel = context.WithCancel(context.Background())

		if err := s.register.AddNode(s.ctx); err != nil {
			panic(fmt.Sprintf("register node error, %v", err))
		}

		go s.register.Heartbeat(s.ctx)
		go s.broadcaster.Subscribe(s.ctx)
	})
}

func (s *Socket) Connect(request *ghttp.Request, clientID string, group ...string) (err error) {
	if clientID == "" {
		return errcode.New(errcode.ErrInvalidRequest, "client id is empty")
	}

	conn, err := s.upgrader.Upgrade(request.Response.Writer, request.Request, nil)
	if err != nil {
		return
	}

	ctx := request.GetCtx()

	client := &Client{
		Socket: s,
		conn:   conn,
		id:     clientID,
		group:  s.connector.groupName(group...),
		input:  make(chan []byte, s.options.InputQueueSize),
		output: make(chan []byte, s.options.OutputQueueSize),
		done:   make(chan struct{}),
	}
	client.ctx, client.cancel = context.WithCancel(context.WithoutCancel(ctx))
	client.Start()

	if err = s.connector.AddClient(ctx, client); err != nil {
		client.Close([]byte("connect failed"))
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
