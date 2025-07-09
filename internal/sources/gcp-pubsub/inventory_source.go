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
	ctx    context.Context
	log    *logrus.Logger
	config *InventorySourceConfig

	pubsub sources.CloseableSource
}

func NewInventorySource(
	ctx context.Context,
	log *logrus.Logger,
	cfg config.GenericConfig,
	pipeline pipeline.IPipelineGroup,
	oasRouter *swagger.Router[fiber.Handler, fiber.Router],
) (sources.CloseableSource, error) {
	config, err := config.GetConfig[*InventorySourceConfig](cfg)
	if err != nil {
		return nil, err
	}

	pipeline.Start(ctx)

	eventBuilder := gcppubsubevents.NewInventoryEventBuilder()

	client, err := internal.New(
		ctx,
		log,
		internal.PubSubConfig{
			ProjectID:          config.ProjectID,
			TopicName:          config.TopicName,
			SubscriptionID:     config.SubscriptionID,
			AckDeadlineSeconds: config.AckDeadlineSeconds,
			CredentialsJSON:    config.CredentialsJSON.String(),
		},
	)
	if err != nil {
		return nil, err
	}

	pubsub, err := newPubSub(ctx, log, pipeline, eventBuilder, client)
	if err != nil {
		return nil, err
	}

	return newInventorySource(ctx, log, config, pipeline, oasRouter, pubsub)
}

func newInventorySource(
	ctx context.Context,
	log *logrus.Logger,
	config *InventorySourceConfig,
	pipeline pipeline.IPipelineGroup,
	oasRouter *swagger.Router[fiber.Handler, fiber.Router],
	pubsub *pubsubConsumer,
) (*InventorySource, error) {
	s := &InventorySource{
		ctx:    ctx,
		log:    log,
		config: config,
		pubsub: pubsub,
	}

	return s, nil
}

func (s *InventorySource) Close() error {
	if err := s.pubsub.Close(); err != nil {
		return err
	}
	return nil
}
