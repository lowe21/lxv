package socket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/lowe21/lxv/pkg/errcode"
	"github.com/lowe21/lxv/pkg/validation"
)

type Client struct {
	*Socket
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
	case <-time.After(c.options.WriteTimeout):
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
		_ = c.connector.DeleteClient(c)
	})
}

func (c *Client) writer() {
	ticker := time.NewTicker(c.options.PingInterval)
	defer func() {
		ticker.Stop()
		c.Close(nil)
	}()

	for {
		select {
		case <-ticker.C:
			if c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(c.options.WriteTimeout)) != nil {
				return
			}
		case message := <-c.output:
			if c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteTimeout)) != nil {
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

	c.conn.SetReadLimit(c.options.MessageMaxSize)
	if c.conn.SetReadDeadline(time.Now().Add(c.options.PongTimeout)) != nil {
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(c.options.PongTimeout))
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
				c.Send(Message(input.Id, "error", errcode.New(errcode.ErrInvalidParam, "message body should be a json object format")))
				continue
			}

			if err := func() (err error) {
				defer func() {
					if exception := recover(); exception != nil {
						g.Log().Errorf(c.ctx, "event handler panic, %+v", exception)
						err = errcode.New(exception)
					} else {
						if err != nil {
							g.Log().Error(c.ctx, gerror.Wrap(err, fmt.Sprintf("%s %s %s", input.Id, input.Event, input.Data)))
						}
					}
				}()

				if err = validation.Validator(c.ctx, input); err != nil {
					return
				}

				event, err := GetEvent(input.Event)
				if err != nil {
					return
				}

				data, err := event.Handler(c.ctx, input)
				if err != nil {
					return
				}

				c.Send(Message(input.Id, input.Event, data))

				return
			}(); err != nil {
				c.Send(Message(input.Id, input.Event, err))
			}
		case <-c.done:
			return
		}
	}
}
