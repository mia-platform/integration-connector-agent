// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

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

	ctx := t.Context()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	require.NoError(t, err)

	ctxPing, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	require.NoError(t, client.Ping(ctxPing, nil))

	coll := client.Database(db).Collection(collection)

	t.Cleanup(func() {
		// here test context is already canceled, so we use a new context
		//nolint: usetesting
		ctx := context.Background()
		err := coll.Drop(ctx)
		require.NoError(t, err)
		err = client.Database(db).Drop(ctx)
		require.NoError(t, err)
		err = client.Disconnect(ctx)
		require.NoError(t, err)
	})

	return coll
}

func RemoveMongoID(docs []map[string]any) []map[string]any {
	results := []map[string]any{}
	for _, doc := range docs {
		newDoc := make(map[string]any)
		for k, v := range doc {
			if k == "_id" {
				continue
			}
			newDoc[k] = v
		}
		results = append(results, newDoc)
	}
	return results
}
