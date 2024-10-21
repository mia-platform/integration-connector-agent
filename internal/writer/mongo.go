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

package writer

import (
	"context"
	"fmt"

	"github.com/mia-platform/data-connector-agent/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

const (
	mongoDBInitializationErrorPrefix = "failed to start mongo writer:"
)

type validateFunc func(context.Context, *mongo.Client) error

// MongoDBConfig contains the configuration needed to connect to a remote MongoDB instance
type MongoDBConfig struct {
	URI        utils.SecretSource `json:"uri"`
	Database   string             `json:"database"`
	Collection string             `json:"collection"`
}

// MongoDBWriter is a concrete implementation of a Writer that will save and delete data from a MongoDB instance.
type MongoDBWriter struct {
	client *mongo.Client

	database   string
	collection string
}

type MongoDBData struct{}

// NewMongoDBWriter will construct a new MongoDB writer and validate the connection parameters via a ping request.
func NewMongoDBWriter(ctx context.Context, config MongoDBConfig) (Writer[*MongoDBData], error) {
	return newMongoDBWriter(ctx, config, func(ctx context.Context, c *mongo.Client) error {
		return c.Ping(ctx, nil)
	})
}

func newMongoDBWriter(ctx context.Context, config MongoDBConfig, validate validateFunc) (Writer[*MongoDBData], error) {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	options, db, collection := mongoClientOptionsFromConfig(config)
	client, err := mongo.Connect(ctxWithCancel, options)
	if err != nil {
		return nil, fmt.Errorf("%s %w", mongoDBInitializationErrorPrefix, err)
	}

	if err := validate(ctxWithCancel, client); err != nil {
		return nil, fmt.Errorf("%s %w", mongoDBInitializationErrorPrefix, err)
	}

	return &MongoDBWriter{
		client:     client,
		database:   db,
		collection: collection,
	}, nil
}

// Write implement the Writer interface. The MongoDBWriter will do an upsert of data using its id as primary key
func (w *MongoDBWriter) Write(ctx context.Context, data *MongoDBData) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := options.FindOneAndReplace()
	opts.SetUpsert(true)

	result := w.client.Database(w.database).
		Collection(w.collection).
		FindOneAndReplace(ctxWithCancel, data.idFilter(), data.bsonData(), opts)
	return result.Err()
}

// Delete implement the Writer interface
func (w *MongoDBWriter) Delete(ctx context.Context, data *MongoDBData) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := options.FindOneAndDelete()
	result := w.client.Database(w.database).
		Collection(w.collection).
		FindOneAndDelete(ctxWithCancel, data.idFilter(), opts)
	return result.Err()
}

// mongoClientOptionsFromConfig return a ClientOptions, database and collection parameters parsed from a
// MongoDBConfig struct.
func mongoClientOptionsFromConfig(config MongoDBConfig) (*options.ClientOptions, string, string) {
	connectionURI := config.URI.Secret()
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

func (d *MongoDBData) idFilter() bson.D {
	//TODO: implement real function
	return bson.D{{Key: "_id", Value: "foo"}}
}

func (d *MongoDBData) bsonData() bson.D {
	//TODO: implement real function
	return bson.D{{Key: "_id", Value: "foo"}}
}
