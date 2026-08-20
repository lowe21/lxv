package socket

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
)

type Register struct {
	socket *Socket
}

func (r *Register) AddNode(ctx context.Context) (err error) {
	return r.socket.redis.SetEX(ctx, r.nodeKey(), r.socket.options.NodeId, int64(r.socket.options.NodeTtl.Seconds()))
}

func (r *Register) RenewNode(ctx context.Context) (err error) {
	if _, err = r.socket.redis.Expire(ctx, r.nodeKey(), int64(r.socket.options.NodeTtl.Seconds())); err != nil {
		return
	}

	return
}

func (r *Register) DeleteNode() (err error) {
	if _, err = r.socket.redis.Del(context.Background(), r.nodeKey()); err != nil {
		return
	}

	return
}

func (r *Register) Heartbeat(ctx context.Context) {
	ticker := time.NewTicker(r.socket.options.NodeHeartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.socket.register.RenewNode(ctx); err != nil {
				g.Log().Errorf(ctx, "heartbeat renew node error: %v", err)
			}

			for _, group := range r.socket.connector.GetGroups() {
				if err := r.socket.connector.RenewClient(ctx, group); err != nil {
					g.Log().Errorf(ctx, "heartbeat renew client error: %v", err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *Register) nodeKey() (key string) {
	return gstr.Join([]string{r.socket.options.RedisKeyPrefix, "node", r.socket.options.NodeId}, ":")
}
