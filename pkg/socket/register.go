package socket

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
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
	key := r.nodeKey()
	ttl := int64(r.options.NodeTTL.Seconds())

	value, err := r.redis.Expire(ctx, key, ttl)
	if err != nil || value > 0 {
		return
	}

	return r.redis.SetEX(ctx, key, r.options.NodeID, ttl)
}

func (r *Register) DeleteNode() (err error) {
	_, err = r.redis.Del(context.Background(), r.nodeKey())

	return
}

func (r *Register) nodeKey() (key string) {
	return r.options.RedisKeyPrefix + ":node:" + r.options.NodeID
}
