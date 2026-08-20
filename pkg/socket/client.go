package socket

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/gogf/gf/v2/util/gconv"

	"github.com/lowe21/lxv/pkg/error_code"
	"github.com/lowe21/lxv/util"
)

type Client struct {
	socket    *Socket
	conn      *websocket.Conn
	id        string
	group     string
	input     chan []byte
	output    chan []byte
	done      chan struct{}
	ctx       context.Context
	cancel    context.CancelFunc
	startOnce sync.Once
	closeOnce sync.Once
}

func (c *Client) Start() {
	c.startOnce.Do(func() {
		go c.writer()
		go c.reader()
		go c.handler()
	})
}

func (c *Client) Send(message []byte) {
	select {
	case <-time.After(c.socket.options.WriteTimeout):
	case c.output <- message:
	case <-c.done:
	}
}

func (c *Client) TrySend(message []byte) {
	select {
	case c.output <- message:
	case <-c.done:
	default:
	}
}

func (c *Client) Close(message []byte) {
	c.closeOnce.Do(func() {
		if c.cancel != nil {
			c.cancel()
		}
		close(c.done)
		if len(message) > 0 {
			_ = c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, string(message)), time.Now().Add(time.Second))
		}
		_ = c.conn.Close()
		_ = c.socket.connector.DeleteClient(c)
	})
}

func (c *Client) writer() {
	ticker := time.NewTicker(c.socket.options.PingIntervalTime)
	defer func() {
		ticker.Stop()
		c.Close(nil)
	}()

	for {
		select {
		case <-ticker.C:
			if c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(c.socket.options.WriteTimeout)) != nil {
				return
			}
		case message := <-c.output:
			if c.conn.SetWriteDeadline(time.Now().Add(c.socket.options.WriteTimeout)) != nil {
				return
			}
			if c.conn.WriteMessage(websocket.TextMessage, message) != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *Client) reader() {
	defer c.Close(nil)

	c.conn.SetReadLimit(c.socket.options.MessageMaxSize)
	if c.conn.SetReadDeadline(time.Now().Add(c.socket.options.PongTimeout)) != nil {
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(c.socket.options.PongTimeout))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		select {
		case c.input <- message:
		case <-c.done:
			return
		}
	}
}

func (c *Client) handler() {
	for {
		select {
		case message := <-c.input:
			input := &Input{}
			if err := gconv.Scan(message, input); err != nil {
				c.Send(Message(input.Id, "error", error_code.New(error_code.InvalidParam, "message body should be a json object format")))
				continue
			}

			if err := util.Validator(c.ctx, input); err != nil {
				c.Send(Message(input.Id, input.Event, err))
				continue
			}

			event, err := GetEvent(input.Event)
			if err != nil {
				c.Send(Message(input.Id, input.Event, err))
				continue
			}

			data, err := event.Handler(c.ctx, input)
			if err != nil {
				c.Send(Message(input.Id, input.Event, err))
				continue
			}

			c.Send(Message(input.Id, input.Event, data))
		case <-c.done:
			return
		}
	}
}
