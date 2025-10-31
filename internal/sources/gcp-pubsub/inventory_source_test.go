// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gcppubsub

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/gcpclient"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook/hmac"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestNewInventorySource(t *testing.T) {
	pg := &pipeline.PipelineGroupMock{}
	log, _ := test.NewNullLogger()

	t.Run("fails on invalid config", func(t *testing.T) {
		_, err := NewInventorySource(t.Context(), log, config.GenericConfig{
			Type: "gcppubsub",
			Raw:  []byte(`{"projectId": "", "topicName": "", "subscriptionId": ""}`),
		}, pg, &swagger.Router[fiber.Handler, fiber.Router]{})

		require.ErrorIs(t, err, config.ErrConfigNotValid)
		require.False(t, pg.StartInvoked)
	})
}

func TestImportWebhook(t *testing.T) {
	pg := &pipeline.Group{}
	log, _ := test.NewNullLogger()

	config := &InventorySourceConfig{
		WebhookPath: "/gcppubsub/import",
	}

	t.Run("exposes import webhook", func(t *testing.T) {
		app, router := testutils.GetTestRouter(t)

		consumer := newInventorySource(t.Context(), log, config, pg, router)
		require.NotNil(t, consumer)
		require.NoError(t, consumer.init(&gcpclient.MockPubSub{}))

		resp, err := app.Test(getWebhookRequest(t, nil))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("does not expose import webhook on missing path from configuration", func(t *testing.T) {
		app, router := testutils.GetTestRouter(t)
		config := &InventorySourceConfig{
			WebhookPath: "",
		}
		consumer := newInventorySource(t.Context(), log, config, pg, router)
		require.NotNil(t, consumer)
		require.NoError(t, consumer.init(&gcpclient.MockPubSub{}))

		resp, err := app.Test(getWebhookRequest(t, nil))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("performs webhook authentication", func(t *testing.T) {
		t.Run("signature is ok", func(t *testing.T) {
			app, router := testutils.GetTestRouter(t)
			config := &InventorySourceConfig{
				WebhookPath: "/gcppubsub/import",
				Authentication: hmac.Authentication{
					HeaderName: "X-Hmac-Signature",
					Secret:     "It's a Secret to Everybody",
				},
			}

			consumer := newInventorySource(t.Context(), log, config, pg, router)
			require.NotNil(t, consumer)
			require.NoError(t, consumer.init(&gcpclient.MockPubSub{}))

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
			config := &InventorySourceConfig{
				WebhookPath: "/gcppubsub/import",
				Authentication: hmac.Authentication{
					HeaderName: "X-Hmac-Signature",
					Secret:     "It's a Secret to Everybody",
				},
			}

			consumer := newInventorySource(t.Context(), log, config, pg, router)
			require.NotNil(t, consumer)
			require.NoError(t, consumer.init(&gcpclient.MockPubSub{}))

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

	t.Run("produces a message for each asset returned by gcp", func(t *testing.T) {
		pg := &pipeline.PipelineGroupMock{
			AssertAddMessage: func(data entities.PipelineEvent) {
				require.NotNil(t, data)
				require.Equal(t, gcpclient.BucketAPI, data.GetType())
			},
		}

		app, router := testutils.GetTestRouter(t)
		client := &gcpclient.MockPubSub{
			ListAssetsResult: []*assetpb.Asset{
				{Name: "//storage.googleapis.com/bucket1", AssetType: gcpclient.BucketAPI},
				{Name: "//storage.googleapis.com/bucket2", AssetType: gcpclient.BucketAPI},
			},
		}

		consumer := newInventorySource(t.Context(), log, config, pg, router)
		require.NotNil(t, consumer)
		require.NoError(t, consumer.init(client))

		resp, err := app.Test(getWebhookRequest(t, nil), -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		require.True(t, client.ListAssetsInvoked())

		require.Len(t, pg.Messages, 2)

		require.Equal(t, gcpclient.BucketAPI, pg.Messages[0].GetType())
		require.Equal(t, entities.Write, pg.Messages[0].Operation())
		require.Equal(t, entities.PkFields{
			entities.PkField{Key: "resourceName", Value: "//storage.googleapis.com/bucket1"},
			entities.PkField{Key: "resourceType", Value: gcpclient.BucketAPI},
		}, pg.Messages[0].GetPrimaryKeys())

		require.Equal(t, gcpclient.BucketAPI, pg.Messages[1].GetType())
		require.Equal(t, entities.Write, pg.Messages[1].Operation())
		require.Equal(t, entities.PkFields{
			entities.PkField{Key: "resourceName", Value: "//storage.googleapis.com/bucket2"},
			entities.PkField{Key: "resourceType", Value: gcpclient.BucketAPI},
		}, pg.Messages[1].GetPrimaryKeys())
	})
}

func getWebhookRequest(t *testing.T, body []byte) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/gcppubsub/import", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
