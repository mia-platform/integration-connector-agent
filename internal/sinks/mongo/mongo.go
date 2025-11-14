// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var (
	mongoTimeout = 5 * time.Second

	ErrMongoInitialization = errors.New("failed to start mongo writer")
)

type validateFunc func(context.Context, *mongo.Client) error

// Writer is a concrete implementation of a Writer that will save and delete data from a MongoDB instance.
type Writer[T entities.PipelineEvent] struct {
	client *mongo.Client

	database      string
	collection    string
	upsertIDField string
	insertOnly    bool
}

// NewMongoDBWriter will construct a new MongoDB writer and validate the connection parameters via a ping request.
func NewMongoDBWriter[T entities.PipelineEvent](ctx context.Context, config *Config) (sinks.Sink[T], error) {
	return newMongoDBWriter[T](ctx, config, func(ctx context.Context, c *mongo.Client) error {
		ctx, cancel := context.WithTimeout(ctx, mongoTimeout)
		defer cancel()
		return c.Ping(ctx, nil)
	})
}

func newMongoDBWriter[T entities.PipelineEvent](ctx context.Context, config *Config, validate validateFunc) (sinks.Sink[T], error) {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	options, db, collection := mongoClientOptionsFromConfig(config)

	client, err := mongo.Connect(ctxWithCancel, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMongoInitialization, err)
	}

	if err := validate(ctxWithCancel, client); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMongoInitialization, err)
	}

	return &Writer[T]{
		client:        client,
		database:      db,
		collection:    collection,
		upsertIDField: "_eventId",
		insertOnly:    config.InsertOnly,
	}, nil
}

func (w *Writer[T]) WriteData(ctx context.Context, data T) error {
	if w.insertOnly {
		return w.Insert(ctx, data)
	}

	switch data.Operation() {
	case entities.Write:
		if err := w.Upsert(ctx, data); err != nil {
			return err
		}
	case entities.Delete:
		if err := w.Delete(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer[T]) Close(ctx context.Context) error {
	return w.client.Disconnect(ctx)
}

func (w *Writer[T]) Insert(ctx context.Context, data T) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := options.InsertOne()
	dataToUpsert, err := data.JSON()
	if err != nil {
		return err
	}

	dataToSave, err := bson.Marshal(dataToUpsert)
	if err != nil {
		return err
	}

	_, err = w.client.Database(w.database).
		Collection(w.collection).
		InsertOne(ctxWithCancel, dataToSave, opts)
	if err != nil {
		return err
	}

	return nil
}

// Write implement the Writer interface. The MongoDBWriter will do an upsert of data using its id as primary key
func (w *Writer[T]) Upsert(ctx context.Context, event T) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := options.Replace()
	opts.SetUpsert(true)

	queryFilter, err := w.idFilter(event)
	if err != nil {
		return err
	}

	parsedData, err := event.JSON()
	if err != nil {
		return err
	}

	w.addPrimaryKeyToData(parsedData, event)

	dataToSave, err := bson.Marshal(parsedData)
	if err != nil {
		return err
	}

	result, err := w.client.Database(w.database).
		Collection(w.collection).
		ReplaceOne(ctxWithCancel, queryFilter, dataToSave, opts)
	if err != nil {
		return err
	}

	if result.UpsertedCount != 1 && result.ModifiedCount != 1 {
		return fmt.Errorf("error upserting data: %d documents upserted", result.UpsertedCount)
	}

	return nil
}

// Delete implement the Writer interface
func (w *Writer[T]) Delete(ctx context.Context, data T) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	queryFilter, err := w.idFilter(data)
	if err != nil {
		return err
	}

	result, err := w.client.Database(w.database).
		Collection(w.collection).
		DeleteOne(ctxWithCancel, queryFilter)
	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("error deleting data: %d documents deleted", result.DeletedCount)
	}

	return nil
}

// mongoClientOptionsFromConfig return a ClientOptions, database and collection parameters parsed from a
// MongoDBConfig struct.
func mongoClientOptionsFromConfig(config *Config) (*options.ClientOptions, string, string) {
	connectionURI := config.URL.String()
	options := options.Client()
	options.ApplyURI(connectionURI)

	database := config.Database
	if len(database) == 0 {
		if cs, err := connstring.ParseAndValidate(connectionURI); err == nil {
			database = cs.Database
		}
	}

	return options, database, config.Collection
}

func (w Writer[T]) idFilter(event T) (bson.D, error) {
	pk := event.GetPrimaryKeys()

	if len(pk) == 0 {
		return bson.D{}, errors.New("missing primary key")
	}

	if len(pk) == 1 {
		return bson.D{{Key: w.upsertIDField, Value: pk[0].Value}}, nil
	}

	filter := bson.D{}
	for _, field := range pk {
		filter = append(filter, bson.E{Key: fmt.Sprintf("%s.%s", w.upsertIDField, field.Key), Value: field.Value})
	}

	return filter, nil
}

func (w Writer[T]) addPrimaryKeyToData(data map[string]any, event T) map[string]any {
	pk := event.GetPrimaryKeys()
	if len(pk) == 1 {
		data[w.upsertIDField] = pk[0].Value
		return data
	}

	data[w.upsertIDField] = pk.Map()
	return data
}
