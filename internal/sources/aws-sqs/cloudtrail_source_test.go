// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package awssqs

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/awsclient"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook/hmac"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	pg := &pipeline.Group{}
	log, _ := test.NewNullLogger()
	_, router := testutils.GetTestRouter(t)

	t.Run("invalid configurations", func(t *testing.T) {
		testCases := []struct {
			config string
		}{
			{config: `{"queueUrl": ""}`},
		}

		for _, tc := range testCases {
			t.Run(tc.config, func(t *testing.T) {
				_, err := NewCloudTrailSource(t.Context(), log, config.GenericConfig{
					Type: "aws-sqs",
					Raw:  []byte(tc.config),
				}, pg, router)
				require.ErrorIs(t, err, config.ErrConfigNotValid)
			})
		}
	})

	t.Run("succeeds with valid config", func(t *testing.T) {
		t.Setenv("MY_SECRET_ENV", "SECRET_VALUE")
		consumer, err := NewCloudTrailSource(t.Context(), log, config.GenericConfig{
			Type: "awssqs",
			Raw:  []byte(`{"queueUrl": "https://something.com","secretAccessKey":{"fromEnv":"MY_SECRET_ENV"},"accessKeyId":"key","region":"us-east-1"}`),
		}, pg, router)

		require.NoError(t, err)
		require.NotNil(t, consumer)
	})
}

func TestImportWebhook(t *testing.T) {
	pg := &pipeline.Group{}
	log, _ := test.NewNullLogger()

	config := &CloudTrailSourceConfig{
		WebhookPath: "/awssqs/import",
	}

	t.Run("exposes import webhook", func(t *testing.T) {
		app, router := testutils.GetTestRouter(t)

		consumer := newCloudTrailSource(t.Context(), log, config, pg, router)
		require.NotNil(t, consumer)
		require.NoError(t, consumer.init(&awsclient.AWSMock{}))

		resp, err := app.Test(getWebhookRequest(t, nil))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("performs webhook authentication", func(t *testing.T) {
		t.Run("signature is ok", func(t *testing.T) {
			app, router := testutils.GetTestRouter(t)
			config := &CloudTrailSourceConfig{
				WebhookPath: "/awssqs/import",
				Authentication: hmac.Authentication{
					HeaderName: "X-Hmac-Signature",
					Secret:     "It's a Secret to Everybody",
				},
			}

			consumer := newCloudTrailSource(t.Context(), log, config, pg, router)
			require.NotNil(t, consumer)
			require.NoError(t, consumer.init(&awsclient.AWSMock{}))

			req := getWebhookRequest(t, nil)
			req.Header.Set("X-Hmac-Signature", "sha256=66a0c074deaa0f489ead6537e0d32f9a344b90bbeda705b6ed45ecd3b413fb40")
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, http.StatusNoContent, resp.StatusCode, "Resp: %s", string(respBody))
		})

		t.Run("signature is NOT ok", func(t *testing.T) {
			app, router := testutils.GetTestRouter(t)
			config := &CloudTrailSourceConfig{
				WebhookPath: "/awssqs/import",
				Authentication: hmac.Authentication{
					HeaderName: "X-Hmac-Signature",
					Secret:     "It's a Secret to Everybody",
				},
			}

			consumer := newCloudTrailSource(t.Context(), log, config, pg, router)
			require.NotNil(t, consumer)
			require.NoError(t, consumer.init(&awsclient.AWSMock{}))

			req := getWebhookRequest(t, nil)
			req.Header.Set("X-Hmac-Signature", "sha256=0000000000000000000000000000000000000000000000000000000000000000")
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Resp: %s", string(respBody))
		})
	})

	t.Run("produces a message for each asset returned by AWS", func(t *testing.T) {
		pg := &pipeline.PipelineGroupMock{
			AssertAddMessage: func(data entities.PipelineEvent) {
				require.NotNil(t, data)
				require.Equal(t, awssqsevents.ImportEventType, data.GetType())
			},
		}

		app, router := testutils.GetTestRouter(t)
		client := &awsclient.AWSMock{
			ListBucketsResult: []*awsclient.Bucket{
				{Name: "bucket1"},
				{Name: "bucket2"},
			},
			ListFunctionsResult: []*awsclient.Function{
				{Name: "function1"},
				{Name: "function2"},
			},
		}

		consumer := newCloudTrailSource(t.Context(), log, config, pg, router)
		require.NotNil(t, consumer)
		require.NoError(t, consumer.init(client))

		resp, err := app.Test(getWebhookRequest(t, nil), -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		require.True(t, client.ListBucketsInvoked())
		require.True(t, client.ListFunctionsInvoked())

		require.Len(t, pg.Messages, 4)

		require.Equal(t, awssqsevents.ImportEventType, pg.Messages[0].GetType())
		require.Equal(t, entities.Write, pg.Messages[0].Operation())
		require.Equal(t, entities.PkFields{
			entities.PkField{Key: "resourceName", Value: "bucket1"},
			entities.PkField{Key: "eventSource", Value: awssqsevents.CloudTrailEventStorageType},
		}, pg.Messages[0].GetPrimaryKeys())

		require.Equal(t, awssqsevents.ImportEventType, pg.Messages[1].GetType())
		require.Equal(t, entities.Write, pg.Messages[1].Operation())
		require.Equal(t, entities.PkFields{
			entities.PkField{Key: "resourceName", Value: "bucket2"},
			entities.PkField{Key: "eventSource", Value: awssqsevents.CloudTrailEventStorageType},
		}, pg.Messages[1].GetPrimaryKeys())

		require.Equal(t, awssqsevents.ImportEventType, pg.Messages[2].GetType())
		require.Equal(t, entities.Write, pg.Messages[2].Operation())
		require.Equal(t, entities.PkFields{
			entities.PkField{Key: "resourceName", Value: "function1"},
			entities.PkField{Key: "eventSource", Value: awssqsevents.CloudTrailEventFunctionType},
		}, pg.Messages[2].GetPrimaryKeys())

		require.Equal(t, awssqsevents.ImportEventType, pg.Messages[3].GetType())
		require.Equal(t, entities.Write, pg.Messages[3].Operation())
		require.Equal(t, entities.PkFields{
			entities.PkField{Key: "resourceName", Value: "function2"},
			entities.PkField{Key: "eventSource", Value: awssqsevents.CloudTrailEventFunctionType},
		}, pg.Messages[3].GetPrimaryKeys())
	})
}

func getWebhookRequest(t *testing.T, body []byte) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/awssqs/import", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
