package socket

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
)

const (
	addClientScript = `
local nodeId = redis.call('HGET', KEYS[1], ARGV[1])
redis.call('HSET', KEYS[1], ARGV[1], ARGV[2])
redis.call('HSET', KEYS[2], ARGV[1], 1)
redis.call('EXPIRE', KEYS[2], ARGV[3])
return nodeId
`
	deleteClientScript = `
local deleted = 0
if redis.call('HGET', KEYS[1], ARGV[1]) == ARGV[2] then
    redis.call('HDEL', KEYS[1], ARGV[1])
    deleted = 1
end
redis.call('HDEL', KEYS[2], ARGV[1])
return deleted
`
	deleteNodeClientScript = `
local deleted = 0
for i = 1, #ARGV, 2 do
    local clientId = ARGV[i]
    local nodeId = ARGV[i + 1]
    if redis.call('HGET', KEYS[1], clientId) == nodeId then
        redis.call('HDEL', KEYS[1], clientId)
        redis.call('HDEL', KEYS[2], clientId)
        deleted = deleted + 1
    end
end
return deleted
`
)

type Connector struct {
	socket  *Socket
	clients map[string]map[string]*Client
	mutex   sync.RWMutex
}

func (c *Connector) GetClient(clientId string, group ...string) (client *Client) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.clients[c.groupName(group...)][clientId]
}

func (c *Connector) GetClients(group ...string) (clients map[string]*Client) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	origClients := c.clients[c.groupName(group...)]
	clients = make(map[string]*Client, len(origClients))
	for id, client := range origClients {
		clients[id] = client
	}

	return
}

func (c *Connector) AddClient(ctx context.Context, client *Client) (err error) {
	data, err := c.socket.redis.Eval(ctx, addClientScript, 2, []string{
		c.groupKey(client.group),
		c.groupNodeKey(c.socket.options.NodeId, client.group),
	}, []any{
		client.id,
		c.socket.options.NodeId,
		int64(c.socket.options.NodeTtl.Seconds()),
	})
	if err != nil {
		return
	}

	nodeId := data.String()
	if nodeId != "" && nodeId != c.socket.options.NodeId {
		if err := c.socket.notifier.CloseClient(ctx, []byte("already connected elsewhere"), nodeId, []string{client.id}, client.group); err != nil {
			g.Log().Errorf(ctx, "close node client error: %v", err)
		}
	}

	c.mutex.Lock()
	clients := c.clients[client.group]
	if clients == nil {
		clients = make(map[string]*Client)
		c.clients[client.group] = clients
	}
	origClient := clients[client.id]
	if origClient != client {
		clients[client.id] = client
	}
	c.mutex.Unlock()

	if origClient != nil && origClient != client {
		origClient.Close([]byte("already connected elsewhere"))
	}

	return
}

func (c *Connector) RenewClients(ctx context.Context, group string) (err error) {
	if _, err = c.socket.redis.Expire(ctx, c.groupNodeKey(c.socket.options.NodeId, group), int64(c.socket.options.NodeTtl.Seconds())); err != nil {
		return
	}

	return
}

func (c *Connector) DeleteClient(client *Client) (err error) {
	c.mutex.Lock()
	deleted := false
	if clients := c.clients[client.group]; clients != nil {
		if clients[client.id] == client {
			delete(clients, client.id)
			if len(clients) == 0 {
				delete(c.clients, client.group)
			}
			deleted = true
		}
	}
	c.mutex.Unlock()

	if deleted {
		if _, err = c.socket.redis.Eval(context.Background(), deleteClientScript, 2, []string{
			c.groupKey(client.group),
			c.groupNodeKey(c.socket.options.NodeId, client.group),
		}, []any{
			client.id,
			c.socket.options.NodeId,
		}); err != nil {
			return
		}
	}

	return
}

func (c *Connector) GetGroups() (groups []string) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	groups = make([]string, 0, len(c.clients))
	for group := range c.clients {
		groups = append(groups, group)
	}

	return
}

func (c *Connector) GetNodeIds(ctx context.Context, clientIds []string, group ...string) (nodeIds []string, err error) {
	data, err := c.socket.redis.HMGet(ctx, c.groupKey(group...), clientIds...)
	if err != nil {
		return
	}

	return data.Strings(), nil
}

func (c *Connector) GetNodeClientIds(ctx context.Context, nodeId string, clientIds []string, group ...string) (activeClientIds []string, err error) {
	data, err := c.socket.redis.HMGet(ctx, c.groupNodeKey(nodeId, group...), clientIds...)
	if err != nil {
		return
	}

	values := data.Strings()

	activeClientIds = make([]string, 0, len(clientIds))
	for index, clientId := range clientIds {
		if index < len(values) && values[index] != "" {
			activeClientIds = append(activeClientIds, clientId)
		}
	}

	return
}

func (c *Connector) DeleteNodeClients(ctx context.Context, nodeId string, clientIds []string, group ...string) (err error) {
	args := make([]any, 0, len(clientIds))
	for _, clientId := range clientIds {
		args = append(args, clientId, nodeId)
	}

	if _, err = c.socket.redis.Eval(ctx, deleteNodeClientScript, 2, []string{
		c.groupKey(group...),
		c.groupNodeKey(nodeId, group...),
	}, args); err != nil {
		return
	}

	return
}

func (c *Connector) groupName(group ...string) (name string) {
	name = c.socket.options.ClientDefaultGroup
	if len(group) > 0 && group[0] != "" {
		name = group[0]
	}

	return
}

func (c *Connector) groupKey(group ...string) (key string) {
	return gstr.Join([]string{c.socket.options.RedisKeyPrefix, "client", c.groupName(group...)}, ":")
}

func (c *Connector) groupNodeKey(nodeId string, group ...string) (key string) {
	return gstr.Join([]string{c.socket.options.RedisKeyPrefix, "client", c.groupName(group...), "node", nodeId}, ":")
}
