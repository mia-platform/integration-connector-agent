package gcppubsub

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	gcppubsubevents "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/events"
	"github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/internal"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestNewInventorySource(t *testing.T) {
	pg := &pipelineGroupMock{}
	log, _ := test.NewNullLogger()

	t.Run("fails on invalid config", func(t *testing.T) {
		_, err := NewInventorySource(t.Context(), log, config.GenericConfig{
			Type: "gcppubsub",
			Raw:  []byte(`{"projectId": "", "topicName": "", "subscriptionId": ""}`),
		}, pg, &swagger.Router[fiber.Handler, fiber.Router]{})

		require.ErrorIs(t, err, config.ErrConfigNotValid)
		require.False(t, pg.startInvoked)
	})

	t.Run("succeeds with valid config", func(t *testing.T) {
		_, router := testutils.GetTestRouter(t)
		consumer, err := NewInventorySource(t.Context(), log, config.GenericConfig{
			Type: "gcppubsub",
			Raw:  []byte(`{"projectId": "test-project", "topicName": "test-topic", "subscriptionId": "test-subscription"}`),
		}, pg, router)

		require.NoError(t, err)
		require.NotNil(t, consumer)
		require.True(t, pg.startInvoked)
	})
}

func TestImportWebhook(t *testing.T) {
	pg := &pipeline.Group{}
	log, _ := test.NewNullLogger()

	config := &InventorySourceConfig{
		ImportTriggerWebhookPath: "/gcppubsub/import",
	}

	t.Run("exposes import webhook", func(t *testing.T) {
		app, router := testutils.GetTestRouter(t)

		consumer := newInventorySource(t.Context(), log, config, pg, router)
		require.NotNil(t, consumer)
		require.NoError(t, consumer.init(&internal.MockPubSub{}))

		resp, err := app.Test(getWebhookRequest(t, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("produces a message for each asset returned by gcp", func(t *testing.T) {
		pg := &pipelineGroupMock{
			assertAddMessage: func(data entities.PipelineEvent) {
				require.NotNil(t, data)
				require.Equal(t, gcppubsubevents.ImportEventType, data.GetType())
			},
		}

		app, router := testutils.GetTestRouter(t)
		client := &internal.MockPubSub{
			ListBucketsResult: []*internal.Bucket{
				{Name: "bucket1"},
				{Name: "bucket2"},
			},
		}

		consumer := newInventorySource(t.Context(), log, config, pg, router)
		require.NotNil(t, consumer)
		require.NoError(t, consumer.init(client))

		resp, err := app.Test(getWebhookRequest(t, nil), -1)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.True(t, client.ListBucketsInvoked())

		require.Len(t, pg.Messages, 2)

		require.Equal(t, gcppubsubevents.ImportEventType, pg.Messages[0].GetType())
		require.Equal(t, entities.Write, pg.Messages[0].Operation())
		require.Equal(t, entities.PkFields{
			entities.PkField{Key: "resourceName", Value: "//storage.googleapis.com/bucket1"},
			entities.PkField{Key: "resourceType", Value: gcppubsubevents.InventoryEventStorageType},
		}, pg.Messages[0].GetPrimaryKeys())

		require.Equal(t, gcppubsubevents.ImportEventType, pg.Messages[1].GetType())
		require.Equal(t, entities.Write, pg.Messages[1].Operation())
		require.Equal(t, entities.PkFields{
			entities.PkField{Key: "resourceName", Value: "//storage.googleapis.com/bucket2"},
			entities.PkField{Key: "resourceType", Value: gcppubsubevents.InventoryEventStorageType},
		}, pg.Messages[1].GetPrimaryKeys())
	})
}

func getWebhookRequest(t *testing.T, body []byte) *http.Request {
	t.Helper()

	req, err := http.NewRequest("POST", "/gcppubsub/import", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}
