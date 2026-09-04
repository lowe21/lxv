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
	SourceNodeID string   `json:"sourceNodeID"`
	TargetNodeID string   `json:"targetNodeID"`
	ClientIDs    []string `json:"clientIDs"`
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
				if broadcast.SourceNodeID == b.options.NodeID && broadcast.TargetNodeID == "" {
					continue
				}
				if broadcast.TargetNodeID != "" && broadcast.TargetNodeID != b.options.NodeID {
					continue
				}

				switch broadcast.Event {
				case eventNotice:
					if len(broadcast.ClientIDs) > 0 {
						for _, clientID := range broadcast.ClientIDs {
							if client := b.connector.GetClient(clientID, broadcast.Group); client != nil {
								client.TrySend(broadcast.Message)
							}
						}
					} else {
						for _, client := range b.connector.GetClients(broadcast.Group) {
							client.TrySend(broadcast.Message)
						}
					}
				case eventCloseClient:
					if len(broadcast.ClientIDs) > 0 {
						for _, clientID := range broadcast.ClientIDs {
							if client := b.connector.GetClient(clientID, broadcast.Group); client != nil {
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

func (b *Broadcaster) Notice(ctx context.Context, message []byte, clientIDs []string, group ...string) (err error) {
	if len(clientIDs) > 0 {
		clientIDsArray := garray.NewStrArrayFrom(clientIDs)
		clientIDsArray.FilterEmpty()
		clientIDsArray.Unique()

		pendingClientIDs := make([]string, 0, clientIDsArray.Len())
		for _, clientID := range clientIDsArray.Slice() {
			if client := b.connector.GetClient(clientID, group...); client != nil {
				client.TrySend(message)
				continue
			}
			pendingClientIDs = append(pendingClientIDs, clientID)
		}

		if len(pendingClientIDs) > 0 {
			targetNodeIDs, err := b.connector.GetNodeIDs(ctx, pendingClientIDs, group...)
			if err != nil {
				return err
			}

			routes := make(map[string][]string, len(targetNodeIDs))
			for index, clientID := range pendingClientIDs {
				if index >= len(targetNodeIDs) {
					continue
				}
				if targetNodeID := targetNodeIDs[index]; targetNodeID != "" {
					routes[targetNodeID] = append(routes[targetNodeID], clientID)
				}
			}

			for targetNodeID, targetClientIDs := range routes {
				activeClientIDs, err := b.connector.GetNodeActiveClientIDs(ctx, targetNodeID, targetClientIDs, group...)
				if err != nil {
					activeClientIDs = targetClientIDs
				}

				staleClientIDs := make([]string, 0, len(targetClientIDs))
				for _, targetClientID := range targetClientIDs {
					isStaled := true
					for _, activeClientID := range activeClientIDs {
						if activeClientID == targetClientID {
							isStaled = false
						}
					}
					if isStaled {
						staleClientIDs = append(staleClientIDs, targetClientID)
					}
				}

				if len(staleClientIDs) > 0 {
					if err := b.connector.DeleteNodeClients(ctx, targetNodeID, staleClientIDs, group...); err != nil {
						g.Log().Errorf(ctx, "delete node clients error, %v", err)
					}
				}

				if len(activeClientIDs) > 0 {
					if _, err = b.redis.GroupPubSub().Publish(ctx, b.channelKey(), &Broadcast{
						Event:        eventNotice,
						SourceNodeID: b.options.NodeID,
						TargetNodeID: targetNodeID,
						ClientIDs:    activeClientIDs,
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
			SourceNodeID: b.options.NodeID,
			Group:        b.connector.groupName(group...),
			Message:      message,
		}); err != nil {
			return
		}
	}

	return
}

func (b *Broadcaster) CloseClient(ctx context.Context, message []byte, nodeID string, clientIDs []string, group ...string) (err error) {
	if nodeID != "" {
		if _, err = b.redis.GroupPubSub().Publish(ctx, b.channelKey(), &Broadcast{
			Event:        eventCloseClient,
			SourceNodeID: b.options.NodeID,
			TargetNodeID: nodeID,
			ClientIDs:    clientIDs,
			Group:        b.connector.groupName(group...),
			Message:      message,
		}); err != nil {
			return
		}
	} else {
		if len(clientIDs) > 0 {
			clientIDsArray := garray.NewStrArrayFrom(clientIDs)
			clientIDsArray.FilterEmpty()
			clientIDsArray.Unique()

			pendingClientIDs := make([]string, 0, clientIDsArray.Len())
			for _, clientID := range clientIDsArray.Slice() {
				if client := b.connector.GetClient(clientID, group...); client != nil {
					client.Close(message)
					continue
				}
				pendingClientIDs = append(pendingClientIDs, clientID)
			}

			if len(pendingClientIDs) > 0 {
				targetNodeIDs, err := b.connector.GetNodeIDs(ctx, pendingClientIDs, group...)
				if err != nil {
					return err
				}

				routes := make(map[string][]string, len(targetNodeIDs))
				for index, clientID := range pendingClientIDs {
					if index >= len(targetNodeIDs) {
						continue
					}
					if targetNodeID := targetNodeIDs[index]; targetNodeID != "" {
						routes[targetNodeID] = append(routes[targetNodeID], clientID)
					}
				}

				for targetNodeID, targetClientIDs := range routes {
					if _, err = b.redis.GroupPubSub().Publish(ctx, b.channelKey(), &Broadcast{
						Event:        eventCloseClient,
						SourceNodeID: b.options.NodeID,
						TargetNodeID: targetNodeID,
						ClientIDs:    targetClientIDs,
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
				SourceNodeID: b.options.NodeID,
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
