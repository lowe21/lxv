package mongodb

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/gogf/gf/v2/container/gmap"
)

var instances = gmap.NewStrAnyMap(true)

func New(name ...string) (mongodb *Mongodb) {
	group := defaultGroup
	if len(name) > 0 && name[0] != "" {
		group = name[0]
	}

	return instances.GetOrSetFuncLock(group, func() any {
		opts := defaultOptions(group)

		client, err := mongo.Connect(options.Client().ApplyURI(opts.Uri).
			SetAppName(opts.AppName).
			SetConnectTimeout(opts.ConnectTimeout).
			SetMaxConnIdleTime(opts.MaxConnIdleTime).
			SetMaxPoolSize(uint64(opts.MaxPoolSize)).
			SetMinPoolSize(uint64(opts.MinPoolSize)).
			SetMaxConnecting(uint64(opts.MaxConnecting)),
		)
		if err != nil {
			panic(err)
		}

		return &Mongodb{
			options: opts,
			client:  client,
		}
	}).(*Mongodb)
}
