package socket

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

type Broadcast struct {
	Event     string   `json:"event"`
	LocalId   string   `json:"localId"`
	NodeId    string   `json:"nodeId"`
	ClientIds []string `json:"clientIds"`
	Group     string   `json:"group"`
	Message   []byte   `json:"message"`
}

type Notifier struct {
	socket *Socket
}

func (n *Notifier) Subscribe(ctx context.Context) {
	for {
		conn, _, err := n.socket.redis.GroupPubSub().Subscribe(ctx, n.channelKey())
		if err != nil {
			g.Log().Errorf(ctx, "subscribe error: %v", err)
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
					g.Log().Errorf(ctx, "receive message error: %v", err)
					continue
				}
				if broadcast.LocalId == n.socket.options.NodeId && broadcast.NodeId == "" {
					continue
				}
				if broadcast.NodeId != "" && broadcast.NodeId != n.socket.options.NodeId {
					continue
				}

				switch broadcast.Event {
				case "notice":
					if len(broadcast.ClientIds) > 0 {
						for _, clientId := range broadcast.ClientIds {
							if client := n.socket.connector.GetClient(clientId, broadcast.Group); client != nil {
								client.TrySend(broadcast.Message)
							}
						}
					} else {
						for _, client := range n.socket.connector.GetClients(broadcast.Group) {
							client.TrySend(broadcast.Message)
						}
					}
				case "closeClient":
					if len(broadcast.ClientIds) > 0 {
						for _, clientId := range broadcast.ClientIds {
							if client := n.socket.connector.GetClient(clientId, broadcast.Group); client != nil {
								client.Close(broadcast.Message)
							}
						}
					} else {
						for _, client := range n.socket.connector.GetClients(broadcast.Group) {
							client.Close(broadcast.Message)
						}
					}
				}
			}
		}

		select {
		case <-time.After(n.socket.options.NodeHeartbeat):
		case <-ctx.Done():
			return
		}
	}
}

func (n *Notifier) Notice(ctx context.Context, message []byte, clientIds []string, group ...string) (err error) {
	if len(clientIds) > 0 {
		clientIdsArray := garray.NewStrArrayFrom(clientIds)
		clientIdsArray.FilterEmpty()
		clientIdsArray.Unique()

		remoteClientIds := make([]string, 0, clientIdsArray.Len())
		for _, clientId := range clientIdsArray.Slice() {
			if client := n.socket.connector.GetClient(clientId, group...); client != nil {
				client.TrySend(message)
				continue
			}
			remoteClientIds = append(remoteClientIds, clientId)
		}

		if len(remoteClientIds) > 0 {
			remoteNodeIds, err := n.socket.connector.GetNodeIds(ctx, remoteClientIds, group...)
			if err != nil {
				return err
			}

			remoteRoutes := make(map[string][]string, len(remoteNodeIds))
			for index, remoteClientId := range remoteClientIds {
				if index >= len(remoteNodeIds) {
					continue
				}
				if remoteNodeId := remoteNodeIds[index]; remoteNodeId != "" {
					remoteRoutes[remoteNodeId] = append(remoteRoutes[remoteNodeId], remoteClientId)
				}
			}

			for remoteNodeId, remoteClientIds := range remoteRoutes {
				activeClientIds, err := n.socket.connector.GetNodeClientIds(ctx, remoteNodeId, remoteClientIds, group...)
				if err != nil {
					g.Log().Errorf(ctx, "get node clients error: %v", err)
					activeClientIds = remoteClientIds
				}

				staleClientIds := make([]string, 0, len(remoteClientIds))
				for _, remoteClientId := range remoteClientIds {
					staled := true
					for _, activeClientId := range activeClientIds {
						if activeClientId == remoteClientId {
							staled = false
						}
					}
					if staled {
						staleClientIds = append(staleClientIds, remoteClientId)
					}
				}

				if len(staleClientIds) > 0 {
					if err := n.socket.connector.DeleteNodeClients(ctx, remoteNodeId, staleClientIds, group...); err != nil {
						g.Log().Errorf(ctx, "delete node clients error: %v", err)
					}
				}

				if len(activeClientIds) > 0 {
					if _, err = n.socket.redis.GroupPubSub().Publish(ctx, n.channelKey(), &Broadcast{
						Event:     "notice",
						LocalId:   n.socket.options.NodeId,
						NodeId:    remoteNodeId,
						ClientIds: activeClientIds,
						Group:     n.socket.connector.groupName(group...),
						Message:   message,
					}); err != nil {
						return err
					}
				}
			}
		}
	} else {
		for _, client := range n.socket.connector.GetClients(group...) {
			client.TrySend(message)
		}

		if _, err = n.socket.redis.GroupPubSub().Publish(ctx, n.channelKey(), &Broadcast{
			Event:   "notice",
			LocalId: n.socket.options.NodeId,
			Group:   n.socket.connector.groupName(group...),
			Message: message,
		}); err != nil {
			return
		}
	}

	return
}

func (n *Notifier) CloseClient(ctx context.Context, message []byte, nodeId string, clientIds []string, group ...string) (err error) {
	if nodeId != "" {
		if _, err = n.socket.redis.GroupPubSub().Publish(ctx, n.channelKey(), &Broadcast{
			Event:     "closeClient",
			LocalId:   n.socket.options.NodeId,
			NodeId:    nodeId,
			ClientIds: clientIds,
			Group:     n.socket.connector.groupName(group...),
			Message:   message,
		}); err != nil {
			return
		}
	} else {
		if len(clientIds) > 0 {
			clientIdsArray := garray.NewStrArrayFrom(clientIds)
			clientIdsArray.FilterEmpty()
			clientIdsArray.Unique()

			remoteClientIds := make([]string, 0, clientIdsArray.Len())
			for _, clientId := range clientIdsArray.Slice() {
				if client := n.socket.connector.GetClient(clientId, group...); client != nil {
					client.Close(message)
					continue
				}
				remoteClientIds = append(remoteClientIds, clientId)
			}

			if len(remoteClientIds) > 0 {
				remoteNodeIds, err := n.socket.connector.GetNodeIds(ctx, remoteClientIds, group...)
				if err != nil {
					return err
				}

				remoteRoutes := make(map[string][]string, len(remoteNodeIds))
				for index, remoteClientId := range remoteClientIds {
					if index >= len(remoteNodeIds) {
						continue
					}
					if remoteNodeId := remoteNodeIds[index]; remoteNodeId != "" {
						remoteRoutes[remoteNodeId] = append(remoteRoutes[remoteNodeId], remoteClientId)
					}
				}

				for remoteNodeId, remoteClientIds := range remoteRoutes {
					if _, err = n.socket.redis.GroupPubSub().Publish(ctx, n.channelKey(), &Broadcast{
						Event:     "closeClient",
						LocalId:   n.socket.options.NodeId,
						NodeId:    remoteNodeId,
						ClientIds: remoteClientIds,
						Group:     n.socket.connector.groupName(group...),
						Message:   message,
					}); err != nil {
						return err
					}
				}
			}
		} else {
			for _, client := range n.socket.connector.GetClients(group...) {
				client.Close(message)
			}

			if _, err = n.socket.redis.GroupPubSub().Publish(ctx, n.channelKey(), &Broadcast{
				Event:   "closeClient",
				LocalId: n.socket.options.NodeId,
				Group:   n.socket.connector.groupName(group...),
				Message: message,
			}); err != nil {
				return
			}
		}
	}

	return
}

func (n *Notifier) channelKey() (key string) {
	return gstr.Join([]string{n.socket.options.RedisKeyPrefix, "channel", n.socket.options.RedisChannel}, ":")
}
