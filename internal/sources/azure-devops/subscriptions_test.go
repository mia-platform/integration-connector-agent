// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package azuredevops

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/google/uuid"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook/basic"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/servicehooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlreadyExists(t *testing.T) {
	t.Parallel()

	testSubscriptions := &[]servicehooks.Subscription{
		{
			Id:          to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
			EventType:   to.Ptr(repositoryCreated),
			PublisherId: to.Ptr(publisherID),
		},
		{
			Id:          to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000002")),
			EventType:   to.Ptr(repositoryDeleted),
			PublisherId: to.Ptr(publisherID),
			ConsumerInputs: &map[string]string{
				"url": "",
			},
		},
		{
			Id:          to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000003")),
			EventType:   to.Ptr(repositoryRenamed),
			PublisherId: to.Ptr(publisherID),
			ConsumerInputs: &map[string]string{
				"url": "",
			},
			PublisherInputs: &map[string]string{
				"projectId": "00000000-0000-0000-0000-00000000000",
			},
		},
	}
	testCases := map[string]struct {
		subscription   servicehooks.Subscription
		expectedResult *servicehooks.Subscription
	}{
		"complete subscription exists": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryRenamed),
				PublisherId: to.Ptr(publisherID),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
				PublisherInputs: &map[string]string{
					"projectId": "00000000-0000-0000-0000-00000000000",
				},
			},
			expectedResult: &servicehooks.Subscription{
				Id:          to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000003")),
				EventType:   to.Ptr(repositoryRenamed),
				PublisherId: to.Ptr(publisherID),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
				PublisherInputs: &map[string]string{
					"projectId": "00000000-0000-0000-0000-00000000000",
				},
			},
		},
		"without publisher inputs exits": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryDeleted),
				PublisherId: to.Ptr(publisherID),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
			},
			expectedResult: &servicehooks.Subscription{
				Id:          to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000002")),
				EventType:   to.Ptr(repositoryDeleted),
				PublisherId: to.Ptr(publisherID),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
			},
		},
		"without consumer & publisher inputs return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryCreated),
				PublisherId: to.Ptr(publisherID),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
			},
		},
		"with different event type return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr("git.pull.created"),
				PublisherId: to.Ptr(publisherID),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
			},
		},
		"with different publisher id return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr("git.pull.deleted"),
				PublisherId: to.Ptr("test"),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
			},
		},
		"with different url return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryRenamed),
				PublisherId: to.Ptr(publisherID),
				ConsumerInputs: &map[string]string{
					"url": "http://example.com",
				},
				PublisherInputs: &map[string]string{
					"projectId": "00000000-0000-0000-0000-00000000000",
				},
			},
		},
		"with nil publisher input return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryDeleted),
				PublisherId: to.Ptr(publisherID),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
				PublisherInputs: &map[string]string{
					"projectId": "00000000-0000-0000-0000-00000000000",
				},
			},
		},
		"with different publisher input return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryRenamed),
				PublisherId: to.Ptr(publisherID),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
				PublisherInputs: &map[string]string{
					"projectId": "00000000-0000-0000-0000-00000000001",
				},
			},
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			existingSubscription := subscriptionAlreadyExists(test.subscription, testSubscriptions)
			assert.Equal(t, test.expectedResult, existingSubscription)
		})
	}
}

func TestSetupSubscriptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		projectsHandler          http.HandlerFunc
		subscriptionQueryHandler http.HandlerFunc
		expectedCreationCount    int
		expectedDeletionCount    int
		expectedError            string
	}{
		"create hooks for one project": {
			projectsHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, err := fmt.Fprint(w, `{"count": 1,"value": [{"id": "6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c","name": "Fabrikam-Fiber-Git","description": "Git projects","url": "https://dev.azure.com/fabrikam/_apis/projects/6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c","state": "wellFormed"}]}`)
				require.NoError(t, err)
			}),
			expectedCreationCount: 3,
		},
		"create hooks for two projects paginated": {
			projectsHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Query().Get("continuationToken") == "" {
					w.Header().Set(azuredevops.HeaderKeyContinuationToken, "1")
					_, err := fmt.Fprint(w, `{"count": 1,"value": [{"id": "6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c","name": "Fabrikam-Fiber-Git","description": "Git projects","url": "https://dev.azure.com/fabrikam/_apis/projects/6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c","state": "wellFormed"}]}`)
					require.NoError(t, err)
					return
				}
				_, err := fmt.Fprint(w, `{"count": 1,"value": [{"id": "281f9a5b-af0d-49b4-a1df-fe6f5e5f84d0","name": "TestGit","url": "https://dev.azure.com/fabrikam/_apis/projects/281f9a5b-af0d-49b4-a1df-fe6f5e5f84d0","state": "wellFormed"}]}`)
				require.NoError(t, err)
			}),
			expectedCreationCount: 6,
		},
		"recreate disable hook for one project and skip an enabled one": {
			projectsHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, err := fmt.Fprint(w, `{"count": 1,"value": [{"id": "6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c","name": "Fabrikam-Fiber-Git","description": "Git projects","url": "https://dev.azure.com/fabrikam/_apis/projects/6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c","state": "wellFormed"}]}`)
				require.NoError(t, err)
			}),
			subscriptionQueryHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				w.Header().Set("Content-Type", "application/json")
				_, err := fmt.Fprint(w, `{"results":[{"id":"00000000-0000-0000-0000-000000000001","eventType":"git.repo.created","publisherId":"tfs","consumerId":"webHooks","consumerActionId":"httpRequest","consumerInputs":{"url":"http://example.com/webhook"},"publisherInputs":{"projectId":"6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c"},"status":"disabled"},{"id":"00000000-0000-0000-0000-000000000002","eventType":"git.repo.deleted","publisherId":"tfs","consumerId":"webHooks","consumerActionId":"httpRequest","consumerInputs":{"url":"http://example.com/webhook"},"publisherInputs":{"projectId":"6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c"},"status":"enabled"}]}`)
				require.NoError(t, err)
			}),
			expectedCreationCount: 2,
			expectedDeletionCount: 1,
		},
		"skip all hooks": {
			projectsHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, err := fmt.Fprint(w, `{"count": 1,"value": [{"id": "6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c","name": "Fabrikam-Fiber-Git","description": "Git projects","url": "https://dev.azure.com/fabrikam/_apis/projects/6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c","state": "wellFormed"}]}`)
				require.NoError(t, err)
			}),
			subscriptionQueryHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				w.Header().Set("Content-Type", "application/json")
				_, err := fmt.Fprint(w, `{"results":[{"id":"00000000-0000-0000-0000-000000000001","eventType":"git.repo.created","publisherId":"tfs","consumerId":"webHooks","consumerActionId":"httpRequest","consumerInputs":{"url":"http://example.com/webhook"},"publisherInputs":{"projectId":"6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c"},"status":"enabled"},{"id":"00000000-0000-0000-0000-000000000002","eventType":"git.repo.deleted","publisherId":"tfs","consumerId":"webHooks","consumerActionId":"httpRequest","consumerInputs":{"url":"http://example.com/webhook"},"publisherInputs":{"projectId":"6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c"},"status":"enabled"}.{"id":"00000000-0000-0000-0000-000000000003","eventType":"git.repo.renamed","publisherId":"tfs","consumerId":"webHooks","consumerActionId":"httpRequest","consumerInputs":{"url":"http://example.com/webhook"},"publisherInputs":{"projectId":"6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c"},"status":"enabled"}]}`)
				require.NoError(t, err)
			}),
		},
		"unparsable continuation token return early": {
			projectsHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set(azuredevops.HeaderKeyContinuationToken, "notanumber")
				_, err := fmt.Fprint(w, `{"count": 1,"value": [{"id": "6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c","name": "Fabrikam-Fiber-Git","description": "Git projects","url": "https://dev.azure.com/fabrikam/_apis/projects/6ce954b1-ce1f-45d1-b94d-e6bf2464ba2c","state": "wellFormed"}]}`)
				require.NoError(t, err)
			}),
			expectedCreationCount: 3,
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			subscriptionsCreated := 0
			subscriptionsDeleted := 0
			server := setupSubscriptionTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet &&
					(r.RequestURI == "/_apis/projects?stateFilter=wellFormed" || r.RequestURI == "/_apis/projects?continuationToken=1&stateFilter=wellFormed"):
					test.projectsHandler.ServeHTTP(w, r)
				case r.Method == http.MethodPost && r.RequestURI == "/_apis/hooks/subscriptionsQuery":
					if test.subscriptionQueryHandler != nil {
						test.subscriptionQueryHandler.ServeHTTP(w, r)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					_, err := fmt.Fprint(w, `{"results":[]}`)
					require.NoError(t, err)
				case r.Method == http.MethodPost && r.RequestURI == "/_apis/hooks/subscriptions":
					defer r.Body.Close()
					subscriptionsCreated++
					w.WriteHeader(http.StatusOK)
					_, err := fmt.Fprint(w, `{}`)
					require.NoError(t, err)
				case r.Method == http.MethodDelete && strings.HasPrefix(r.RequestURI, "/_apis/hooks/subscriptions/"):
					subscriptionsDeleted++
					w.WriteHeader(http.StatusOK)
				default:
					t.Fatalf("unexpected request: %s %s", r.Method, r.RequestURI)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			connection := azuredevops.NewPatConnection(server.URL, "pat")
			err := setupSubscriptions(t.Context(), connection, &Config{
				WebhookHost: "http://example.com",
				Configuration: webhook.Configuration[*basic.Authentication]{
					WebhookPath: "/webhook",
					Authentication: &basic.Authentication{
						Username: "user",
						Secret:   "secret",
					},
				},
			})

			if len(test.expectedError) > 0 {
				assert.ErrorContains(t, err, test.expectedError)
				return
			}

			require.Equal(t, test.expectedCreationCount, subscriptionsCreated)
			require.Equal(t, test.expectedDeletionCount, subscriptionsDeleted)
		})
	}
}

func setupSubscriptionTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	testServer := httptest.NewServer(nil)
	testServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Basic OnBhdA==", r.Header.Get("Authorization"))
		switch {
		case r.Method == http.MethodOptions && r.RequestURI == "/_apis":
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprint(w, `{"value":[{"id":"e81700f7-3be2-46de-8624-2eb35882fcaa","area":"Location","resourceName":"ResourceAreas","routeTemplate":"_apis/{resource}/{areaId}","resourceVersion":1,"minVersion":"3.2","maxVersion":"7.2","releasedVersion":"0.0"},{"id":"603fe2ac-9723-48b9-88ad-09305aa6c6e1","area":"core","resourceName":"projects","routeTemplate":"_apis/{resource}/{*projectId}","resourceVersion":4,"minVersion":"1.0","maxVersion":"7.2","releasedVersion":"7.1"},{"id":"c7c3c1cf-9e05-4c0d-a425-a0f922c2c6ed","area":"hooks","resourceName":"subscriptionsQuery","routeTemplate":"_apis/{area}/{resource}","resourceVersion":1,"minVersion":"1.0","maxVersion":"7.2","releasedVersion":"7.1"},{"id":"fc50d02a-849f-41fb-8af1-0a5216103269","area":"hooks","resourceName":"subscriptions","routeTemplate":"_apis/{area}/{resource}/{subscriptionId}","resourceVersion":1,"minVersion":"1.0","maxVersion":"7.2","releasedVersion":"7.1"}],"count":5}`)
			require.NoError(t, err)
			return
		case r.Method == http.MethodGet && r.RequestURI == "/_apis/ResourceAreas":
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprintf(w, `{"value":[{"id":"79134c72-4a58-4b42-976c-04e7115f32bf","name":"core","locationUrl":"%[1]s/"}],"count":1}`, testServer.URL)
			require.NoError(t, err)
		default:
			handler.ServeHTTP(w, r)
		}
	})
	return testServer
}
