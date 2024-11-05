// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mongo

import (
	"context"
	"errors"
	"fmt"

	"github.com/mia-platform/data-connector-agent/internal/entities"
	"github.com/mia-platform/data-connector-agent/internal/writer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var (
	ErrMongoInitialization = errors.New("failed to start mongo writer")
)

type validateFunc func(context.Context, *mongo.Client) error

// Config contains the configuration needed to connect to a remote MongoDB instance
type Config struct {
	URI         string
	Database    string
	Collection  string
	OutputModel map[string]any
}

// Writer is a concrete implementation of a Writer that will save and delete data from a MongoDB instance.
type Writer[T entities.PipelineEvent] struct {
	client *mongo.Client

	database    string
	collection  string
	outputModel map[string]any
}

// NewMongoDBWriter will construct a new MongoDB writer and validate the connection parameters via a ping request.
func NewMongoDBWriter[T entities.PipelineEvent](ctx context.Context, config Config) (writer.Writer[T], error) {
	return newMongoDBWriter[T](ctx, config, func(ctx context.Context, c *mongo.Client) error {
		return c.Ping(ctx, nil)
	})
}

func newMongoDBWriter[T entities.PipelineEvent](ctx context.Context, config Config, validate validateFunc) (writer.Writer[T], error) {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	options, db, collection := mongoClientOptionsFromConfig(config)

	client, err := mongo.Connect(ctxWithCancel, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMongoInitialization, err)
	}

	if err := validate(ctxWithCancel, client); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMongoInitialization, err)
	}

	return &Writer[T]{
		client:      client,
		database:    db,
		collection:  collection,
		outputModel: config.OutputModel,
	}, nil
}

// Write implement the Writer interface. The MongoDBWriter will do an upsert of data using its id as primary key
func (w *Writer[T]) Write(ctx context.Context, data T) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := options.FindOneAndReplace()
	opts.SetUpsert(true)

	queryFilter, err := w.idFilter(data)
	if err != nil {
		return err
	}

	result := w.client.Database(w.database).
		Collection(w.collection).
		FindOneAndReplace(ctxWithCancel, queryFilter, w.bsonData(data), opts)
	return result.Err()
}

// Delete implement the Writer interface
func (w *Writer[T]) Delete(ctx context.Context, data T) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	queryFilter, err := w.idFilter(data)
	if err != nil {
		return err
	}

	opts := options.FindOneAndDelete()
	result := w.client.Database(w.database).
		Collection(w.collection).
		FindOneAndDelete(ctxWithCancel, queryFilter, opts)
	return result.Err()
}

func (w *Writer[T]) OutputModel() map[string]any {
	return w.outputModel
}

// mongoClientOptionsFromConfig return a ClientOptions, database and collection parameters parsed from a
// MongoDBConfig struct.
func mongoClientOptionsFromConfig(config Config) (*options.ClientOptions, string, string) {
	connectionURI := config.URI
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

func (w Writer[T]) idFilter(data T) (bson.D, error) {
	id := data.GetID()
	if id == "" {
		return bson.D{}, fmt.Errorf("id is empty")
	}
	return bson.D{{Key: "_id", Value: data.GetID()}}, nil
}

func (w Writer[T]) bsonData(_ T) bson.D {
	//TODO: implement real function
	return bson.D{{Key: "_id", Value: "foo"}}
}
