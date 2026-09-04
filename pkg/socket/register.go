package socket

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
)

type Register struct {
	*Socket
}

func (r *Register) Heartbeat(ctx context.Context) {
	ticker := time.NewTicker(r.options.NodeHeartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.register.RenewNode(ctx); err != nil {
				g.Log().Errorf(ctx, "heartbeat renew node error, %v", err)
			}

			for _, group := range r.connector.GetGroups() {
				if err := r.connector.RenewClients(ctx, group); err != nil {
					g.Log().Errorf(ctx, "heartbeat renew clients error, %v", err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *Register) AddNode(ctx context.Context) (err error) {
	return r.redis.SetEX(ctx, r.nodeKey(), r.options.NodeID, int64(r.options.NodeTTL.Seconds()))
}

func (r *Register) RenewNode(ctx context.Context) (err error) {
	if _, err = r.redis.Expire(ctx, r.nodeKey(), int64(r.options.NodeTTL.Seconds())); err != nil {
		return
	}

	return
}

func (r *Register) DeleteNode() (err error) {
	if _, err = r.redis.Del(context.Background(), r.nodeKey()); err != nil {
		return
	}

	return
}

func (r *Register) nodeKey() (key string) {
	return gstr.Join([]string{r.options.RedisKeyPrefix, "node", r.options.NodeID}, ":")
}
