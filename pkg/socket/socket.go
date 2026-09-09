package socket

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/lowe21/lxv/pkg/errcode"
)

type Socket struct {
	options     *Options
	upgrader    websocket.Upgrader
	redis       *gredis.Redis
	register    *Register
	connector   *Connector
	broadcaster *Broadcaster
	started     bool
	ctx         context.Context
	cancel      context.CancelFunc
	mutex       sync.RWMutex
	wg          sync.WaitGroup
}

func (s *Socket) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.started {
		return
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())

	if err := s.register.AddNode(s.ctx); err != nil {
		g.Log().Errorf(s.ctx, "register node error, %v", err)
	}

	s.wg.Add(2)
	go func() {
		defer s.wg.Done()
		s.register.Heartbeat(s.ctx)
	}()
	go func() {
		defer s.wg.Done()
		s.broadcaster.Subscribe(s.ctx)
	}()

	s.started = true
}

func (s *Socket) Connect(request *ghttp.Request, clientID string, group ...string) (err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.started {
		return errcode.New(errcode.ErrInvalidRequest, "socket is not started")
	}

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
		token:  guid.S(),
		group:  s.connector.groupName(group...),
		input:  make(chan []byte, s.options.InputQueueSize),
		output: make(chan []byte, s.options.OutputQueueSize),
		done:   make(chan struct{}),
	}
	client.ctx, client.cancel = context.WithCancel(context.WithoutCancel(ctx))

	if err = s.connector.AddClient(ctx, client); err != nil {
		client.Close([]byte("connect failed"))
	} else {
		client.Start()
		client.Send(Message(client.id, "connect", "connect succeed"))
	}

	return
}

func (s *Socket) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.started {
		s.started = false
	} else {
		return
	}

	if s.cancel != nil {
		s.cancel()
	}

	s.wg.Wait()

	for _, group := range s.connector.GetGroups() {
		for _, client := range s.connector.GetClients(group) {
			client.Close([]byte("shutdown"))
		}
	}

	if err := s.register.DeleteNode(); err != nil {
		g.Log().Errorf(s.ctx, "unregister client error, %v", err)
	}
}
