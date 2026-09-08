package socket

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	addClientScript = `
local nodeID = redis.call('HGET', KEYS[1], ARGV[1])
local token = ''
if nodeID and nodeID ~= ARGV[2] then
    token = redis.call('HGET', ARGV[5] .. nodeID, ARGV[1]) or ''
end
redis.call('HSET', KEYS[1], ARGV[1], ARGV[2])
redis.call('HSET', KEYS[2], ARGV[1], ARGV[3])
redis.call('EXPIRE', KEYS[2], ARGV[4])
return {nodeID or '', token}
`

	renewClientsScript = `
local ttl = ARGV[1]
for i = 2, #ARGV, 2 do
    redis.call('HSETNX', KEYS[1], ARGV[i], ARGV[i + 1])
end
redis.call('EXPIRE', KEYS[1], ttl)
return 1
`

	deleteClientScript = `
local token = redis.call('HGET', KEYS[2], ARGV[1])
if token and token == ARGV[3] then
    if redis.call('HGET', KEYS[1], ARGV[1]) == ARGV[2] then
        redis.call('HDEL', KEYS[1], ARGV[1])
    end
    redis.call('HDEL', KEYS[2], ARGV[1])
    return 1
end
return 0
`

	deleteNodeClientScript = `
local deleted = 0
for i = 1, #ARGV, 2 do
    local clientID = ARGV[i]
    local nodeID = ARGV[i + 1]
    if redis.call('HGET', KEYS[1], clientID) == nodeID and redis.call('HEXISTS', KEYS[2], clientID) == 0 then
        redis.call('HDEL', KEYS[1], clientID)
        deleted = deleted + 1
    end
end
return deleted
`
)

type Connector struct {
	*Socket
	clients map[string]map[string]*Client
	mutex   sync.RWMutex
}

func (c *Connector) GetClient(clientID string, group ...string) (client *Client) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.clients[c.groupName(group...)][clientID]
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
	c.mutex.Lock()
	data, err := c.redis.Eval(ctx, addClientScript, 2, []string{
		c.groupKey(client.group),
		c.groupNodeKey(c.options.NodeID, client.group),
	}, []any{
		client.id,
		c.options.NodeID,
		client.token,
		int64(c.options.NodeTTL.Seconds()),
		c.groupNodeKey("", client.group),
	})
	if err != nil {
		c.mutex.Unlock()
		return
	}

	values := data.Slice()
	nodeID := gconv.String(values[0])
	token := gconv.String(values[1])

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

	if nodeID != "" && nodeID != c.options.NodeID && token != "" {
		if err := c.broadcaster.CloseClient(ctx, []byte("already connected elsewhere"), nodeID, []string{client.id}, map[string]string{
			client.id: token,
		}, client.group); err != nil {
			g.Log().Errorf(ctx, "close client error, %v", err)
		}
	}

	if origClient != nil && origClient != client {
		origClient.Close([]byte("already connected elsewhere"))
	}

	return
}

func (c *Connector) RenewClients(ctx context.Context, group string) (err error) {
	key := c.groupNodeKey(c.options.NodeID, group)
	ttl := int64(c.options.NodeTTL.Seconds())

	result, err := c.redis.Expire(ctx, key, ttl)
	if err != nil {
		return
	}
	if result > 0 {
		return
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	clients := c.clients[c.groupName(group)]
	if len(clients) > 0 {
		args := make([]any, 0, len(clients)*2+1)
		args = append(args, ttl)
		for _, client := range clients {
			args = append(args, client.id, client.token)
		}

		if _, err = c.redis.Eval(ctx, renewClientsScript, 1, []string{key}, args); err != nil {
			return
		}
	}

	return
}

func (c *Connector) DeleteClient(client *Client) (err error) {
	c.mutex.Lock()
	isDeleted := false
	if clients := c.clients[client.group]; clients != nil {
		if clients[client.id] == client {
			delete(clients, client.id)
			if len(clients) == 0 {
				delete(c.clients, client.group)
			}
			isDeleted = true
		}
	}
	c.mutex.Unlock()

	if isDeleted {
		if _, err = c.redis.Eval(context.Background(), deleteClientScript, 2, []string{
			c.groupKey(client.group),
			c.groupNodeKey(c.options.NodeID, client.group),
		}, []any{
			client.id,
			c.options.NodeID,
			client.token,
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

func (c *Connector) GetNodeIDs(ctx context.Context, clientIDs []string, group ...string) (nodeIDs []string, err error) {
	data, err := c.redis.HMGet(ctx, c.groupKey(group...), clientIDs...)
	if err != nil {
		return
	}

	return data.Strings(), nil
}

func (c *Connector) GetNodeActiveClientIDs(ctx context.Context, nodeID string, clientIDs []string, group ...string) (activeClientIDs []string, err error) {
	data, err := c.redis.HMGet(ctx, c.groupNodeKey(nodeID, group...), clientIDs...)
	if err != nil {
		return
	}

	values := data.Strings()

	activeClientIDs = make([]string, 0, len(clientIDs))
	for index, clientID := range clientIDs {
		if index < len(values) && values[index] != "" {
			activeClientIDs = append(activeClientIDs, clientID)
		}
	}

	return
}

func (c *Connector) DeleteNodeClients(ctx context.Context, nodeID string, clientIDs []string, group ...string) (err error) {
	args := make([]any, 0, len(clientIDs)*2)
	for _, clientID := range clientIDs {
		args = append(args, clientID, nodeID)
	}

	if _, err = c.redis.Eval(ctx, deleteNodeClientScript, 2, []string{
		c.groupKey(group...),
		c.groupNodeKey(nodeID, group...),
	}, args); err != nil {
		return
	}

	return
}

func (c *Connector) groupName(group ...string) (name string) {
	name = c.options.DefaultClientGroup
	if len(group) > 0 && group[0] != "" {
		name = group[0]
	}

	return
}

func (c *Connector) groupKey(group ...string) (key string) {
	return gstr.Join([]string{c.options.RedisKeyPrefix, "client", c.groupName(group...)}, ":")
}

func (c *Connector) groupNodeKey(nodeID string, group ...string) (key string) {
	return gstr.Join([]string{c.options.RedisKeyPrefix, "client", c.groupName(group...), "node", nodeID}, ":")
}
