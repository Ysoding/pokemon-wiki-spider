package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DBer interface {
	Insert(table TableData) error
	InsertMany(table TableData) error
}

type TableData struct {
	TableName string
	Data      []interface{}
}

type MongoDB struct {
	client *mongo.Client
	options
}

func (db *MongoDB) Insert(table TableData) error {
	collection := db.client.Database(db.dbName).Collection(table.TableName)
	_, err := collection.InsertOne(context.TODO(), table.Data[0])
	return err
}

func (db *MongoDB) InsertMany(table TableData) error {
	collection := db.client.Database(db.dbName).Collection(table.TableName)
	_, err := collection.InsertMany(context.TODO(), table.Data)
	return err
}

func (db *MongoDB) Find(tableName string, filter interface{}, result interface{}) error {
	collection := db.client.Database(db.dbName).Collection(tableName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, result); err != nil {
		return err
	}

	return nil
}

func New(opts ...Option) (*MongoDB, error) {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	s := &MongoDB{options: options}

	if err := s.OpenDB(); err != nil {
		return nil, err
	}

	return s, nil
}

func (m *MongoDB) OpenDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	m.logger.Info("mongodb init success")
	client, err := mongo.Connect(ctx, mongooptions.Client().ApplyURI(m.uri))
	if err != nil {
		return err
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	m.logger.Info("mongodb ping success")
	m.client = client
	return nil
}
