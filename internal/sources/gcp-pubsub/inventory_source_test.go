package gcppubsub

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestNewInventorySource(t *testing.T) {
	pg := &pipeline.Group{}
	log, _ := test.NewNullLogger()

	t.Run("fails on invalid config", func(t *testing.T) {
		_, err := NewInventorySource(t.Context(), log, config.GenericConfig{
			Type: "gcppubsub",
			Raw:  []byte(`{"projectId": "", "topicName": "", "subscriptionId": ""}`),
		}, pg, &swagger.Router[fiber.Handler, fiber.Router]{})

		require.ErrorIs(t, err, config.ErrConfigNotValid)
	})

	t.Run("succeeds with valid config", func(t *testing.T) {
		consumer, err := NewInventorySource(t.Context(), log, config.GenericConfig{
			Type: "gcppubsub",
			Raw:  []byte(`{"projectId": "test-project", "topicName": "test-topic", "subscriptionId": "test-subscription"}`),
		}, pg, &swagger.Router[fiber.Handler, fiber.Router]{})

		require.NoError(t, err)
		require.NotNil(t, consumer)
	})
}

func TestImportWebhook(t *testing.T) {
	pg := &pipeline.Group{}
	log, _ := test.NewNullLogger()

	config := &InventorySourceConfig{}
	// 	ImportTriggerWebhookPath: "/gcppubsub/import",
	// }

	t.Run("exposes import webhook", func(t *testing.T) {
		app, router := testutils.GetTestRouter(t)
		// e := &eventBuilderMock{}
		pubsub := &pubsubConsumer{}
		consumer, err := newInventorySource(t.Context(), log, config, pg, router, pubsub)
		require.NoError(t, err)
		require.NotNil(t, consumer)

		resp, err := app.Test(getWebhookRequest(t, nil))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// t.Run("produces a message for each asset returned by gcp", func(t *testing.T) {
	// 	pg := &gcppubsubevents.PipelineGroupMock{
	// 		AssertAddMessage: func(data entities.PipelineEvent) {
	// 			require.NotNil(t, data)
	// 			require.Equal(t, "some-type", data.GetType())
	// 		},
	// 	}

	// 	app, router := testutils.GetTestRouter(t)
	// 	client := &internal.MockPubSub{
	// 		ListBucketsResult: []*internal.Bucket{
	// 			{Name: "bucket1"},
	// 		},
	// 	}

	// 	consumer, err := new(t.Context(), log, config, pg, router, client)
	// 	require.NoError(t, err)
	// 	require.NotNil(t, consumer)

	// 	resp, err := app.Test(getWebhookRequest(t, nil))
	// 	require.NoError(t, err)
	// 	require.Equal(t, http.StatusOK, resp.StatusCode)

	// 	require.True(t, client.ListBucketsInvoked())

	// 	require.Len(t, pg.Messages, 1)
	// 	require.Equal(t, "import", pg.Messages[0].GetType())
	// })
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
