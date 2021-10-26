package services

import (
	"context"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoOnce sync.Once
var myMongo *Mongo

type Mongo struct {
	client *mongo.Client
	db     string
}

func GetMongo() *Mongo {
	mongoOnce.Do(func() {
		myMongo = &Mongo{}
	})
	return myMongo
}

func (mg *Mongo) cursorToArray(ctx context.Context, cur *mongo.Cursor) ([]interface{}, error) {
	var results []interface{}
	err := cur.All(ctx, &results)
	if err != nil {
		return nil, err
	}
	err = cur.Close(ctx)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (mg *Mongo) Connect(ctx context.Context) {
	config := GetConfig()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.Mongo.URL))
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	mg.client = client
	mg.db = config.Mongo.DB
}

func (mg *Mongo) Disconnect(ctx context.Context) error {
	return mg.client.Disconnect(ctx)
}

func (mg *Mongo) Database() *mongo.Database {
	return mg.client.Database(mg.db)
}

func (mg *Mongo) Find(ctx context.Context, col string, filter interface{}, opts ...*options.FindOptions) ([]interface{}, error) {
	cur, err := mg.Database().Collection(col).Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return mg.cursorToArray(ctx, cur)
}

func (mg *Mongo) UpdateMany(ctx context.Context, col string, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return mg.Database().Collection(col).UpdateMany(ctx, filter, update, opts...)
}

func (mg *Mongo) BulkWrite(ctx context.Context, col string, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	return mg.Database().Collection(col).BulkWrite(ctx, models, opts...)
}

func (mg *Mongo) CreateIndex(ctx context.Context, col string, model mongo.IndexModel, opts ...*options.CreateIndexesOptions) (string, error) {
	return mg.Database().Collection(col).Indexes().CreateOne(ctx, model, opts...)
}

func (mg *Mongo) ListIndexes(ctx context.Context, col string, opts ...*options.ListIndexesOptions) ([]primitive.D, error) {
	cur, err := mg.Database().Collection(col).Indexes().List(ctx, opts...)
	if err != nil {
		return nil, err
	}
	var results []primitive.D
	err = cur.All(ctx, &results)
	if err != nil {
		return nil, err
	}
	err = cur.Close(ctx)
	if err != nil {
		return nil, err
	}
	return results, nil
}
