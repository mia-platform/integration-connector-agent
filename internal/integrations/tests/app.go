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

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/server"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"

	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

type setupServerConfig struct {
	configPath string
}

func setupApp(tb testing.TB, setupCfg setupServerConfig) (*fiber.App, string, string) {
	tb.Helper()

	mongoURL, db := testutils.GenerateMongoURL(tb)
	tb.Setenv("INTEGRATION_TEST_MONGO_URL", mongoURL)

	envVars, err := env.ParseAsWithOptions[config.EnvironmentVariables](env.Options{
		Environment: map[string]string{
			"CONFIGURATION_PATH":     setupCfg.configPath,
			"DELAY_SHUTDOWN_SECONDS": "0",
		},
	})
	require.NoError(tb, err)
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)

	cfg, err := config.LoadServiceConfiguration(envVars.ConfigurationPath)
	require.NoError(tb, err)

	ctx := context.Background()
	app, err := server.NewApp(ctx, envVars, log, cfg)
	require.NoError(tb, err)

	return app, mongoURL, db
}

func findAllDocuments(t *testing.T, coll *mongo.Collection, expectedResults []map[string]any) {
	t.Helper()

	require.Eventuallyf(t, func() bool {
		n, err := coll.CountDocuments(context.Background(), map[string]any{})
		require.NoError(t, err)
		return n == int64(len(expectedResults))
	}, 1*time.Second, 10*time.Millisecond, "invalid document length")

	ctx := context.Background()
	docs, err := coll.Find(ctx, map[string]any{})
	require.NoError(t, err)
	results := []map[string]any{}
	err = docs.All(ctx, &results)
	require.NoError(t, err)

	ok := assert.Eventuallyf(t, func() bool {
		return assert.ObjectsAreEqual(expectedResults, testutils.RemoveMongoID(results))
	}, 1*time.Second, 10*time.Millisecond, "results not corrects")
	// This is only needed to show the diffs in case of failure
	if !ok {
		require.Equal(t, expectedResults, testutils.RemoveMongoID(results))
	}
}
