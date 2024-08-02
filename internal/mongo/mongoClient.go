package mongocon

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	Addr   *string
	Client *mongo.Client
}

func NewMongoClient(addr string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+ addr).SetCompressors([]string{"zstd"}))
	if err != nil {
		return nil, err
	}
	return &MongoClient{
		Addr:   &addr,
		Client: client,
	}, nil
}

func (mc *MongoClient) InsertData(data map[string]string) error {
	collection := mc.Client.Database("cert").Collection("certs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func (mc *MongoClient) Close() error {
	err := mc.Client.Disconnect(context.Background())
	if err != nil {
		return err
	}
	return nil
}