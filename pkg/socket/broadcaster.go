package socket

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	eventNotice      = "notice"
	eventCloseClient = "closeClient"
)

type Broadcast struct {
	Event        string   `json:"event"`
	SourceNodeId string   `json:"sourceNodeId"`
	TargetNodeId string   `json:"targetNodeId"`
	ClientIds    []string `json:"clientIds"`
	Group        string   `json:"group"`
	Message      []byte   `json:"message"`
}

type Broadcaster struct {
	*Socket
}

func (b *Broadcaster) Subscribe(ctx context.Context) {
	for {
		conn, _, err := b.redis.GroupPubSub().Subscribe(ctx, b.channelKey())
		if err != nil {
			g.Log().Errorf(ctx, "subscribe error, %v", err)
		} else {
			done := make(chan struct{})
			go func() {
				select {
				case <-done:
				case <-ctx.Done():
					_ = conn.Close(nil)
				}
			}()

			for {
				message, err := conn.ReceiveMessage(ctx)
				if err != nil {
					close(done)
					_ = conn.Close(nil)
					break
				}

				broadcast := &Broadcast{}
				if err = gconv.Scan(message.Payload, broadcast); err != nil {
					g.Log().Errorf(ctx, "receive message error, %v", err)
					continue
				}
				if broadcast.SourceNodeId == b.options.NodeId && broadcast.TargetNodeId == "" {
					continue
				}
				if broadcast.TargetNodeId != "" && broadcast.TargetNodeId != b.options.NodeId {
					continue
				}

				switch broadcast.Event {
				case eventNotice:
					if len(broadcast.ClientIds) > 0 {
						for _, clientId := range broadcast.ClientIds {
							if client := b.connector.GetClient(clientId, broadcast.Group); client != nil {
								client.TrySend(broadcast.Message)
							}
						}
					} else {
						for _, client := range b.connector.GetClients(broadcast.Group) {
							client.TrySend(broadcast.Message)
						}
					}
				case eventCloseClient:
					if len(broadcast.ClientIds) > 0 {
						for _, clientId := range broadcast.ClientIds {
							if client := b.connector.GetClient(clientId, broadcast.Group); client != nil {
								client.Close(broadcast.Message)
							}
						}
					} else {
						for _, client := range b.connector.GetClients(broadcast.Group) {
							client.Close(broadcast.Message)
						}
					}
				}
			}
		}

		select {
		case <-time.After(b.options.NodeHeartbeat):
		case <-ctx.Done():
			return
		}
	}
}

func (b *Broadcaster) Notice(ctx context.Context, message []byte, clientIds []string, group ...string) (err error) {
	if len(clientIds) > 0 {
		clientIdsArray := garray.NewStrArrayFrom(clientIds)
		clientIdsArray.FilterEmpty()
		clientIdsArray.Unique()

		pendingClientIds := make([]string, 0, clientIdsArray.Len())
		for _, clientId := range clientIdsArray.Slice() {
			if client := b.connector.GetClient(clientId, group...); client != nil {
				client.TrySend(message)
				continue
			}
			pendingClientIds = append(pendingClientIds, clientId)
		}

		if len(pendingClientIds) > 0 {
			targetNodeIds, err := b.connector.GetNodeIds(ctx, pendingClientIds, group...)
			if err != nil {
				return err
			}

			routes := make(map[string][]string, len(targetNodeIds))
			for index, clientId := range pendingClientIds {
				if index >= len(targetNodeIds) {
					continue
				}
				if targetNodeId := targetNodeIds[index]; targetNodeId != "" {
					routes[targetNodeId] = append(routes[targetNodeId], clientId)
				}
			}

			for targetNodeId, targetClientIds := range routes {
				activeClientIds, err := b.connector.GetNodeActiveClientIds(ctx, targetNodeId, targetClientIds, group...)
				if err != nil {
					activeClientIds = targetClientIds
				}

				staleClientIds := make([]string, 0, len(targetClientIds))
				for _, targetClientId := range targetClientIds {
					isStaled := true
					for _, activeClientId := range activeClientIds {
						if activeClientId == targetClientId {
							isStaled = false
						}
					}
					if isStaled {
						staleClientIds = append(staleClientIds, targetClientId)
					}
				}

				if len(staleClientIds) > 0 {
					if err := b.connector.DeleteNodeClients(ctx, targetNodeId, staleClientIds, group...); err != nil {
						g.Log().Errorf(ctx, "delete node clients error, %v", err)
					}
				}

				if len(activeClientIds) > 0 {
					if _, err = b.redis.GroupPubSub().Publish(ctx, b.channelKey(), &Broadcast{
						Event:        eventNotice,
						SourceNodeId: b.options.NodeId,
						TargetNodeId: targetNodeId,
						ClientIds:    activeClientIds,
						Group:        b.connector.groupName(group...),
						Message:      message,
					}); err != nil {
						return err
					}
				}
			}
		}
	} else {
		for _, client := range b.connector.GetClients(group...) {
			client.TrySend(message)
		}

		if _, err = b.redis.GroupPubSub().Publish(ctx, b.channelKey(), &Broadcast{
			Event:        eventNotice,
			SourceNodeId: b.options.NodeId,
			Group:        b.connector.groupName(group...),
			Message:      message,
		}); err != nil {
			return
		}
	}

	return
}

func (b *Broadcaster) CloseClient(ctx context.Context, message []byte, nodeId string, clientIds []string, group ...string) (err error) {
	if nodeId != "" {
		if _, err = b.redis.GroupPubSub().Publish(ctx, b.channelKey(), &Broadcast{
			Event:        eventCloseClient,
			SourceNodeId: b.options.NodeId,
			TargetNodeId: nodeId,
			ClientIds:    clientIds,
			Group:        b.connector.groupName(group...),
			Message:      message,
		}); err != nil {
			return
		}
	} else {
		if len(clientIds) > 0 {
			clientIdsArray := garray.NewStrArrayFrom(clientIds)
			clientIdsArray.FilterEmpty()
			clientIdsArray.Unique()

			pendingClientIds := make([]string, 0, clientIdsArray.Len())
			for _, clientId := range clientIdsArray.Slice() {
				if client := b.connector.GetClient(clientId, group...); client != nil {
					client.Close(message)
					continue
				}
				pendingClientIds = append(pendingClientIds, clientId)
			}

			if len(pendingClientIds) > 0 {
				targetNodeIds, err := b.connector.GetNodeIds(ctx, pendingClientIds, group...)
				if err != nil {
					return err
				}

				routes := make(map[string][]string, len(targetNodeIds))
				for index, clientId := range pendingClientIds {
					if index >= len(targetNodeIds) {
						continue
					}
					if targetNodeId := targetNodeIds[index]; targetNodeId != "" {
						routes[targetNodeId] = append(routes[targetNodeId], clientId)
					}
				}

				for targetNodeId, targetClientIds := range routes {
					if _, err = b.redis.GroupPubSub().Publish(ctx, b.channelKey(), &Broadcast{
						Event:        eventCloseClient,
						SourceNodeId: b.options.NodeId,
						TargetNodeId: targetNodeId,
						ClientIds:    targetClientIds,
						Group:        b.connector.groupName(group...),
						Message:      message,
					}); err != nil {
						return err
					}
				}
			}
		} else {
			for _, client := range b.connector.GetClients(group...) {
				client.Close(message)
			}

			if _, err = b.redis.GroupPubSub().Publish(ctx, b.channelKey(), &Broadcast{
				Event:        eventCloseClient,
				SourceNodeId: b.options.NodeId,
				Group:        b.connector.groupName(group...),
				Message:      message,
			}); err != nil {
				return
			}
		}
	}

	return
}

func (b *Broadcaster) channelKey() (key string) {
	return gstr.Join([]string{b.options.RedisKeyPrefix, "channel", b.options.RedisChannel}, ":")
}
