package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/gogf/gf/v2/errors/gerror"
)

type Mongodb struct {
	options *Options
	client  *mongo.Client
}

func (m *Mongodb) ObjectId(hex string) (id bson.ObjectID, err error) {
	return bson.ObjectIDFromHex(hex)
}

func (m *Mongodb) Count(ctx context.Context, collection string, filter any) (count int64, err error) {
	return m.client.Database(m.options.Database).Collection(collection).CountDocuments(ctx, filter)
}

func (m *Mongodb) Find(ctx context.Context, collection string, filter, pointer any) (err error) {
	if err = m.client.Database(m.options.Database).Collection(collection).FindOne(ctx, filter).Decode(pointer); err != nil {
		if gerror.Is(err, mongo.ErrNoDocuments) {
			err = nil
		}
	}

	return
}

func (m *Mongodb) Insert(ctx context.Context, collection string, document any) (id any, err error) {
	result, err := m.client.Database(m.options.Database).Collection(collection).InsertOne(ctx, document)
	if err != nil {
		return
	}

	return result.InsertedID, nil
}

func (m *Mongodb) Update(ctx context.Context, collection string, filter, document any) (affected int64, err error) {
	result, err := m.client.Database(m.options.Database).Collection(collection).UpdateOne(ctx, filter, document)
	if err != nil {
		return
	}

	return result.ModifiedCount, nil
}

func (m *Mongodb) Delete(ctx context.Context, collection string, filter any) (affected int64, err error) {
	result, err := m.client.Database(m.options.Database).Collection(collection).DeleteOne(ctx, filter)
	if err != nil {
		return
	}

	return result.DeletedCount, nil
}
