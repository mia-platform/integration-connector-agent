package gcppubsub

import (
	"context"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources"
	gcppubsubevents "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/events"
	"github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/internal"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type InventorySourceConfig struct {
	ProjectID          string              `json:"projectId"`
	TopicName          string              `json:"topicName"`
	SubscriptionID     string              `json:"subscriptionId"`
	AckDeadlineSeconds int                 `json:"ackDeadlineSeconds,omitempty"`
	CredentialsJSON    config.SecretSource `json:"credentialsJson,omitempty"`
}

func (c *InventorySourceConfig) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("projectId must be provided")
	}
	if c.TopicName == "" {
		return fmt.Errorf("topicName must be provided")
	}
	if c.SubscriptionID == "" {
		return fmt.Errorf("subscriptionId must be provided")
	}

	return nil
}

type InventorySource struct {
	pubsub internal.PubSub
	config *InventorySourceConfig
}

func NewInventorySource(
	ctx context.Context,
	log *logrus.Logger,
	cfg config.GenericConfig,
	pipeline pipeline.IPipelineGroup,
	oasRouter *swagger.Router[fiber.Handler, fiber.Router],
) (sources.CloseableSource, error) {
	eventBuilder := gcppubsubevents.NewInventoryEventBuilder()

	config, err := config.GetConfig[*InventorySourceConfig](cfg)
	if err != nil {
		return nil, err
	}

	pubsub, err := newPubSub(&pubSubConfig{
		ctx:                ctx,
		log:                log,
		ProjectID:          config.ProjectID,
		TopicName:          config.TopicName,
		SubscriptionID:     config.SubscriptionID,
		AckDeadlineSeconds: config.AckDeadlineSeconds,
		CredentialsJSON:    config.CredentialsJSON.String(),
	}, pipeline, eventBuilder)
	if err != nil {
		return nil, err
	}

	s := &InventorySource{
		pubsub: pubsub.client,
		config: config,
	}

	return s, nil
}

// func newInventorySource()

func (s *InventorySource) Close() error {
	if err := s.pubsub.Close(); err != nil {
		return err
	}
	return nil
}
