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

package testutils

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const localhostMongoDB = "localhost:27017"

func GenerateMongoURL(tb testing.TB) (string, string) {
	tb.Helper()

	host, ok := os.LookupEnv("MONGO_HOST_CI")
	if !ok {
		host = localhostMongoDB
	}

	db := RandomString(tb, 10)
	tb.Logf("Generated db: %s", db)

	return "mongodb://" + host + "/" + db, db
}

func MongoCollection(t *testing.T, mongoURL, collection, db string) *mongo.Collection {
	t.Helper()

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	require.NoError(t, err)

	ctxPing, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	require.NoError(t, client.Ping(ctxPing, nil))
	t.Cleanup(func() {
		err := client.Database(db).Drop(ctx)
		require.NoError(t, err)
		err = client.Disconnect(ctx)
		require.NoError(t, err)
	})

	return client.Database(db).Collection(collection)
}

func RemoveMongoID(docs []map[string]any) []map[string]any {
	results := []map[string]any{}
	for _, doc := range docs {
		delete(doc, "_id")
		results = append(results, doc)
	}
	return results
}
