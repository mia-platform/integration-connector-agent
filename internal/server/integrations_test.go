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

//go:build integration
// +build integration

package server

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"
	"github.com/mia-platform/integration-connector-agent/internal/sources"

	swagger "github.com/davidebianchi/gswagger"
	oasfiber "github.com/davidebianchi/gswagger/support/fiber"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestSetupWriters(t *testing.T) {
	ctx := t.Context()

	testCases := map[string]struct {
		writers config.Sinks

		expectError string
	}{
		"unsupported writer type": {
			writers: config.Sinks{
				{
					Type: "unsupported",
				},
			},
			expectError: "unsupported writer type: unsupported",
		},
		"multiple writers": {
			writers: config.Sinks{
				getFakeWriter(t),
				getFakeWriter(t),
			},
		},
		"crud service writer": {
			writers: config.Sinks{
				config.GenericConfig{
					Type: sinks.CRUDService,
					Raw:  []byte(`{"url":"https://some-url.com"}`),
				},
			},
		},
		"crud service writer fail for invalid configuration": {
			writers: config.Sinks{
				config.GenericConfig{
					Type: sinks.CRUDService,
					Raw:  []byte(`{"url":""}`),
				},
			},
			expectError: "error setting up writer: configuration not valid: URL not set in CRUD service sink configuration",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			log, _ := test.NewNullLogger()
			w, err := setupSinks(ctx, log, tc.writers)

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
		// Use jsonCfg to test the real test cases, because the GenericConfig to work correctly needs to be unmarshaled
		jsonCfg string

		expectError          string
		expectedIntegrations int
	}{
		"more than 1 writers not supported": {
			cfg: config.Configuration{
				Integrations: []config.Integration{
					{
						Source: config.GenericConfig{
							Type: sources.Jira,
						},
						Pipelines: []config.Pipeline{
							{
								Sinks: config.Sinks{
									getFakeWriter(t),
									getFakeWriter(t),
								},
							},
						},
					},
				},
			},
			expectError: "only 1 writer is supported, now there are 2",
		},
		"multiple integration sources": {
			cfg: config.Configuration{
				Integrations: []config.Integration{
					{
						Source:    config.GenericConfig{Type: sources.Jira, Raw: []byte(`{}`)},
						Pipelines: []config.Pipeline{{Sinks: config.Sinks{getFakeWriter(t)}}},
					},
					{
						Source:    config.GenericConfig{Type: sources.Jira, Raw: []byte(`{}`)},
						Pipelines: []config.Pipeline{{Sinks: config.Sinks{getFakeWriter(t)}}},
					},
				},
			},
			expectedIntegrations: 2,
		},
		"unsupported writer type": {
			cfg: config.Configuration{
				Integrations: []config.Integration{
					{
						Source: config.GenericConfig{
							Type: sources.Jira,
						},
						Pipelines: []config.Pipeline{
							{
								Sinks: config.Sinks{
									{Type: "unsupported"},
								},
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
						Source: config.GenericConfig{
							Type: "test",
						},
						Pipelines: []config.Pipeline{
							{
								Sinks: config.Sinks{
									getFakeWriter(t),
								},
							},
						},
					},
				},
			},
		},
		"unsupported integration type": {
			cfg: config.Configuration{
				Integrations: []config.Integration{
					{
						Source: config.GenericConfig{
							Type: "unsupported",
						},
						Pipelines: []config.Pipeline{
							{
								Sinks: config.Sinks{
									getFakeWriter(t),
								},
							},
						},
					},
				},
			},
			expectError: "unsupported integration type: unsupported",
		},
		"jira integration type": {
			jsonCfg: `{"integrations":[{"source":{"type":"jira"},"pipelines":[{"sinks":[{"type":"fake","raw":{}}]}]}]}`,
		},
		"console integration type": {
			jsonCfg: `{"integrations":[{"source":{"type":"console"},"pipelines":[{"sinks":[{"type":"fake","raw":{}}]}]}]}`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			log, _ := test.NewNullLogger()
			router := getRouter(t)

			integrations, err := setupIntegrations(ctx, log, &tc.cfg, router)
			if tc.expectError != "" {
				require.EqualError(t, err, tc.expectError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, integrations)
				require.Equal(t, tc.expectedIntegrations, len(integrations))
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

func getFakeWriter(t *testing.T) config.GenericConfig {
	t.Helper()

	return config.GenericConfig{
		Type: sinks.Fake,
		Raw:  []byte(`{}`),
	}
}
