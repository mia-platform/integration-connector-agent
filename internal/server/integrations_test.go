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

package server

import (
	"context"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"
	integration "github.com/mia-platform/integration-connector-agent/internal/sources"

	swagger "github.com/davidebianchi/gswagger"
	oasfiber "github.com/davidebianchi/gswagger/support/fiber"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestSetupWriters(t *testing.T) {
	ctx := context.Background()

	testCases := map[string]struct {
		writers []config.Writer

		expectError string
	}{
		"unsupported writer type": {
			writers: []config.Writer{
				{
					Type: "unsupported",
				},
			},
			expectError: "unsupported writer type: unsupported",
		},
		"multiple writers": {
			writers: []config.Writer{
				getFakeWriter(t),
				getFakeWriter(t),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w, err := setupWriters(ctx, tc.writers)

			if tc.expectError != "" {
				require.EqualError(t, err, tc.expectError)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.writers), len(w))
			}
		})
	}
}

func TestSetupIntegrations(t *testing.T) {
	testCases := map[string]struct {
		cfg config.Configuration

		expectError string
	}{
		"more than 1 writers not supported": {
			cfg: config.Configuration{
				Integrations: []config.Integration{
					{
						Type: integration.Jira,
						Writers: []config.Writer{
							getFakeWriter(t),
							getFakeWriter(t),
						},
					},
				},
			},
			expectError: "only 1 writer is supported, now there are 2 for integration jira",
		},
		"unsupported writer type": {
			cfg: config.Configuration{
				Integrations: []config.Integration{
					{
						Type: integration.Jira,
						Writers: []config.Writer{
							{
								Type: "unsupported",
							},
						},
					},
				},
			},
			expectError: "unsupported writer type: unsupported",
		},
		"setup test integration": {
			cfg: config.Configuration{
				Integrations: []config.Integration{
					{
						Type: "test",
						Writers: []config.Writer{
							getFakeWriter(t),
						},
					},
				},
			},
		},
		"unsupported integration type": {
			cfg: config.Configuration{
				Integrations: []config.Integration{
					{
						Type: "unsupported",
						Writers: []config.Writer{
							getFakeWriter(t),
						},
					},
				},
			},
			expectError: "unsupported integration type",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			log, _ := test.NewNullLogger()
			router := getRouter(t)

			err := setupIntegrations(ctx, log, &tc.cfg, router)
			if tc.expectError != "" {
				require.EqualError(t, err, tc.expectError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func getRouter(t *testing.T) *swagger.Router[fiber.Handler, fiber.Router] {
	t.Helper()

	app := fiber.New()
	router, err := swagger.NewRouter(oasfiber.NewRouter(app), swagger.Options{
		Openapi: &openapi3.T{
			OpenAPI: "3.1.0",
			Info: &openapi3.Info{
				Title:   "Test",
				Version: "test-version",
			},
		},
	})
	require.NoError(t, err)

	return router
}

func getFakeWriter(t *testing.T) config.Writer {
	t.Helper()

	return config.Writer{
		Type: sinks.Fake,
		Raw:  []byte(`{}`),
	}
}
