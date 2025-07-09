package gcppubsub

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"

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
