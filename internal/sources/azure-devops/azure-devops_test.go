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

package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/google/uuid"
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/processors"
	fakewriter "github.com/mia-platform/integration-connector-agent/internal/sinks/fake"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/gofiber/fiber/v2"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		config        *Config
		expectedError error
	}{
		"valid": {
			config: &Config{
				ImportWebhookPath:              "/webhook",
				AzureDevOpsOrganizationURL:     "https://dev.azure.com/myorg",
				AzureDevOpsPersonalAccessToken: config.SecretSource("pat"),
				WebhookHost:                    "https://example.com",
			},
		},
		"valid with missing webhook path": {
			config: &Config{
				AzureDevOpsOrganizationURL:     "https://dev.azure.com/myorg",
				AzureDevOpsPersonalAccessToken: config.SecretSource("pat"),
				WebhookHost:                    "https://example.com",
			},
		},
		"missing organization path": {
			config: &Config{
				ImportWebhookPath:              "/webhook",
				AzureDevOpsPersonalAccessToken: config.SecretSource("pat"),
				WebhookHost:                    "https://example.com",
			},
			expectedError: ErrMissingRequiredField,
		},
		"missing PAT": {
			config: &Config{
				ImportWebhookPath:          "/webhook",
				AzureDevOpsOrganizationURL: "https://dev.azure.com/myorg",
				WebhookHost:                "https://example.com",
			},
			expectedError: ErrMissingRequiredField,
		},
		"missing host": {
			config: &Config{
				ImportWebhookPath:              "/webhook",
				AzureDevOpsPersonalAccessToken: config.SecretSource("pat"),
				AzureDevOpsOrganizationURL:     "https://dev.azure.com/myorg",
			},
			expectedError: ErrMissingRequiredField,
		},
		"unparsable host": {
			config: &Config{
				ImportWebhookPath:              "/webhook",
				AzureDevOpsPersonalAccessToken: config.SecretSource("pat"),
				AzureDevOpsOrganizationURL:     "https://dev.azure.com/myorg",
				WebhookHost:                    "://wrong.host",
			},
			expectedError: ErrInvalidHost,
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			err := test.config.Validate()
			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
				return
			}

			assert.NoError(t, err)
		})
	}
}

var repository1 = git.GitRepository{
	Id:   to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
	Name: to.Ptr("repo"),
	Url:  to.Ptr("https://example.com/organization/00000000-0000-0000-0000-000000000000/_apis/git/repositories/00000000-0000-0000-0000-000000000000"),
	Project: &core.TeamProjectReference{
		Id:         to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
		Name:       to.Ptr("project"),
		Url:        to.Ptr("https://example.com/organization/_apis/projects/00000000-0000-0000-0000-000000000000"),
		State:      &core.ProjectStateValues.WellFormed,
		Revision:   to.Ptr(uint64(27)),
		Visibility: &core.ProjectVisibilityValues.Private,
		LastUpdateTime: &azuredevops.Time{
			Time: time.Now(),
		},
	},
	DefaultBranch:   to.Ptr("refs/heads/main"),
	Size:            to.Ptr(uint64(10000)),
	RemoteUrl:       to.Ptr("https://organization@example.com/organization/project/_git/repo"),
	SshUrl:          to.Ptr("git@example.com:organization/project/repo"),
	WebUrl:          to.Ptr("https://example.com/organization/project/_git/repo"),
	IsDisabled:      to.Ptr(false),
	IsInMaintenance: to.Ptr(false),
}

var repository2 = git.GitRepository{
	Id:   to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
	Name: to.Ptr("repo2"),
	Url:  to.Ptr("https://example.com/organization/00000000-0000-0000-0000-000000000001/_apis/git/repositories/00000000-0000-0000-0000-000000000001"),
	Project: &core.TeamProjectReference{
		Id:         to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
		Name:       to.Ptr("project"),
		Url:        to.Ptr("https://example.com/organization/_apis/projects/00000000-0000-0000-0000-000000000001"),
		State:      &core.ProjectStateValues.WellFormed,
		Revision:   to.Ptr(uint64(27)),
		Visibility: &core.ProjectVisibilityValues.Private,
		LastUpdateTime: &azuredevops.Time{
			Time: time.Now(),
		},
	},
	DefaultBranch:   to.Ptr("refs/heads/main"),
	Size:            to.Ptr(uint64(10000)),
	RemoteUrl:       to.Ptr("https://organization@example.com/organization/project/_git/repo2"),
	SshUrl:          to.Ptr("git@example.com:organization/project/repo2"),
	WebUrl:          to.Ptr("https://example.com/organization/project/_git/repo2"),
	IsDisabled:      to.Ptr(false),
	IsInMaintenance: to.Ptr(false),
}

func TestAddSourceToRouter(t *testing.T) {
	t.Parallel()

	repo1Data, err := json.Marshal(repository1)
	require.NoError(t, err)

	repo2Data, err := json.Marshal(repository2)
	require.NoError(t, err)

	testCases := map[string]struct {
		handler       http.Handler
		expectedCalls fakewriter.Calls
		expectedError string
	}{
		"get only one repository": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err = fmt.Fprintf(w, `{"value":[%s],"count":1}`, string(repo1Data))
				require.NoError(t, err)
			}),
			expectedCalls: fakewriter.Calls{
				{
					Data: &entities.Event{
						PrimaryKeys: entities.PkFields{
							{
								Key:   "repositoryId",
								Value: "00000000-0000-0000-0000-000000000000",
							},
							{
								Key:   "type",
								Value: "repository",
							},
						},
						Type:          "azure-devops-repository",
						OperationType: entities.Write,
						OriginalRaw:   repo1Data,
					},
					Operation: entities.Write,
				},
			},
		},
		"get multiple repository": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err = fmt.Fprintf(w, `{"value":[%s,%s],"count":2}`, string(repo1Data), string(repo2Data))
				require.NoError(t, err)
			}),
			expectedCalls: fakewriter.Calls{
				{
					Data: &entities.Event{
						PrimaryKeys: entities.PkFields{
							{
								Key:   "repositoryId",
								Value: "00000000-0000-0000-0000-000000000000",
							},
							{
								Key:   "type",
								Value: "repository",
							},
						},
						Type:          "azure-devops-repository",
						OperationType: entities.Write,
						OriginalRaw:   repo1Data,
					},
					Operation: entities.Write,
				},
				{
					Data: &entities.Event{
						PrimaryKeys: entities.PkFields{
							{
								Key:   "repositoryId",
								Value: "00000000-0000-0000-0000-000000000001",
							},
							{
								Key:   "type",
								Value: "repository",
							},
						},
						Type:          "azure-devops-repository",
						OperationType: entities.Write,
						OriginalRaw:   repo2Data,
					},
					Operation: entities.Write,
				},
			},
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			server := testServer(t, test.handler)
			defer server.Close()

			app, sink, pg := setup(t, server.URL)
			defer pg.Close()

			response, err := app.Test(httptest.NewRequest(http.MethodPost, "/webhook", nil))
			if len(test.expectedError) > 0 {
				assert.ErrorContains(t, err, test.expectedError)
				return
			}

			require.NoError(t, err)
			defer response.Body.Close()
			assert.Equal(t, http.StatusNoContent, response.StatusCode)
			assert.Eventually(t, func() bool {
				return len(sink.Calls()) == len(test.expectedCalls)
			}, 1*time.Second, 10*time.Millisecond)

			assert.Equal(t, test.expectedCalls, sink.Calls())
		})
	}
}

func setup(t *testing.T, serverURL string) (*fiber.App, *fakewriter.Writer, pipeline.IPipelineGroup) {
	t.Helper()

	logger, _ := logrustest.NewNullLogger()
	app, router := testutils.GetTestRouter(t)

	proc := &processors.Processors{}
	sink := fakewriter.New(nil)
	p, err := pipeline.New(logger, proc, sink)
	require.NoError(t, err)
	pg := pipeline.NewGroup(logger, p)

	rawConfig := json.RawMessage(fmt.Sprintf(`{
	"type": "azure-devops",
	"importWebhookPath": "/webhook",
	"azureDevOpsOrganizationUrl": "%s",
	"azureDevOpsPersonalAccessToken": {
		"fromFile": "testdata/pat-file"
	},
	"webhookHost": "https://example.com"
}`, serverURL))

	cfg := new(config.GenericConfig)
	json.Unmarshal(rawConfig, cfg)
	err = AddSourceToRouter(context.WithoutCancel(t.Context()), *cfg, pg, router)
	require.NoError(t, err)

	return app, sink, pg
}

func testServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	testServer := httptest.NewServer(nil)
	testServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Basic OnBhdA==", r.Header.Get("Authorization"))
		fmt.Println("Request:", r.Method, r.RequestURI)
		switch {
		case r.Method == http.MethodOptions && r.RequestURI == "/_apis":
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprint(w, `{"value":[{"id":"225f7195-f9c7-4d14-ab28-a83f7ff77e1f","area":"git","resourceName":"repositories","routeTemplate":"{project}/_apis/{area}/{resource}/{repositoryId}","resourceVersion":2,"minVersion":"1.0","maxVersion":"7.2","releasedVersion":"7.1"},{"id":"e81700f7-3be2-46de-8624-2eb35882fcaa","area":"Location","resourceName":"ResourceAreas","routeTemplate":"_apis/{resource}/{areaId}","resourceVersion":1,"minVersion":"3.2","maxVersion":"7.2","releasedVersion":"0.0"},{"id":"603fe2ac-9723-48b9-88ad-09305aa6c6e1","area":"core","resourceName":"projects","routeTemplate":"_apis/{resource}/{*projectId}","resourceVersion":4,"minVersion":"1.0","maxVersion":"7.2","releasedVersion":"7.1"},{"id":"c7c3c1cf-9e05-4c0d-a425-a0f922c2c6ed","area":"hooks","resourceName":"subscriptionsQuery","routeTemplate":"_apis/{area}/{resource}","resourceVersion":1,"minVersion":"1.0","maxVersion":"7.2","releasedVersion":"7.1"}],"count":4}`)
			require.NoError(t, err)
			return
		case r.Method == http.MethodGet && r.RequestURI == "/_apis/ResourceAreas":
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprintf(w, `{"value":[{"id": "4e080c62-fa21-4fbc-8fef-2a10a2b38049","name": "git","locationUrl": "%s/"},{"id":"79134c72-4a58-4b42-976c-04e7115f32bf","name":"core","locationUrl":"%[1]s/"}],"count":2}`, testServer.URL)
			require.NoError(t, err)
		case r.Method == http.MethodGet && r.RequestURI == "/_apis/projects?stateFilter=wellFormed":
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprintf(w, `{"count":0,"value":[]}`)
			require.NoError(t, err)
		default:
			handler.ServeHTTP(w, r)
		}
	})

	return testServer
}
