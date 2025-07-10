package gcppubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources"
	gcppubsubevents "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/events"
	"github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/gcpclient"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

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

	WebhookPath    string       `json:"webhookPath,omitempty"`
	Authentication webhook.HMAC `json:"authentication"`
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
	ctx      context.Context
	log      *logrus.Logger
	config   *InventorySourceConfig
	pipeline pipeline.IPipelineGroup

	gcp    gcpclient.GCP
	pubsub sources.CloseableSource
	router *swagger.Router[fiber.Handler, fiber.Router]
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

	client, err := gcpclient.New(
		ctx,
		log,
		gcpclient.GCPConfig{
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

	s := newInventorySource(ctx, log, config, pipeline, oasRouter)
	if err := s.init(client); err != nil {
		return nil, fmt.Errorf("failed to initialize inventory source: %w", err)
	}
	return s, nil
}

func newInventorySource(
	ctx context.Context,
	log *logrus.Logger,
	config *InventorySourceConfig,
	pipeline pipeline.IPipelineGroup,
	oasRouter *swagger.Router[fiber.Handler, fiber.Router],
) *InventorySource {
	return &InventorySource{
		ctx:      ctx,
		log:      log,
		config:   config,
		pipeline: pipeline,
		router:   oasRouter,
	}
}

func (s *InventorySource) init(client gcpclient.GCP) error {
	s.pipeline.Start(s.ctx)

	s.gcp = client

	eventBuilder := gcppubsubevents.NewInventoryEventBuilder[gcppubsubevents.InventoryEvent]()
	pubsub, err := newPubSub(s.ctx, s.log, s.pipeline, eventBuilder, s.gcp)
	if err != nil {
		return fmt.Errorf("failed to create pubsub consumer: %w", err)
	}
	s.pubsub = pubsub

	if s.config.WebhookPath != "" {
		s.log.WithField("webhookPath", s.config.WebhookPath).Info("Registering import webhook")
		if err := s.registerImportWebhook(); err != nil {
			return fmt.Errorf("failed to register import webhook: %w", err)
		}
	}

	return nil
}

func (s *InventorySource) Close() error {
	if err := s.pubsub.Close(); err != nil {
		return err
	}
	return nil
}

func (s *InventorySource) registerImportWebhook() error {
	apiPath := s.config.WebhookPath

	_, err := s.router.AddRoute(http.MethodPost, apiPath, s.webhookHandler, swagger.Definitions{})
	return err
}

func (s *InventorySource) webhookHandler(c *fiber.Ctx) error {
	if err := s.config.Authentication.CheckSignature(c); err != nil {
		s.log.WithError(err).Error("error validating webhook request")
		return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
	}

	buckets, err := s.gcp.ListBuckets(c.UserContext())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to list buckets: " + err.Error()))
	}

	eventBuilder := gcppubsubevents.NewInventoryEventBuilder[gcppubsubevents.InventoryImportEvent]()

	for _, bucket := range buckets {
		importEvent := gcppubsubevents.InventoryImportEvent{
			AssetName: bucket.AssetName(),
			Type:      gcppubsubevents.InventoryEventStorageType,
		}
		data, err := json.Marshal(importEvent)
		if err != nil {
			s.log.WithField("bucketName", bucket.Name).WithError(err).Warn("failed to create import event data for bucket")
			continue
		}

		event, err := eventBuilder.GetPipelineEvent(s.ctx, data)
		if err != nil {
			s.log.WithField("bucketName", bucket.Name).WithError(err).Warn("failed to create import event for bucket")
			continue
		}

		s.pipeline.AddMessage(event)
	}

	functions, err := s.gcp.ListFunctions(c.UserContext())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to list functions: " + err.Error()))
	}
	for _, function := range functions {
		importEvent := gcppubsubevents.InventoryImportEvent{
			AssetName: function.AssetName(),
			Type:      gcppubsubevents.InventoryEventFunctionType,
		}
		data, err := json.Marshal(importEvent)
		if err != nil {
			s.log.WithField("functionName", function.Name).WithError(err).Warn("failed to create import event data for function")
			continue
		}

		event, err := eventBuilder.GetPipelineEvent(s.ctx, data)
		if err != nil {
			s.log.WithField("functionName", function.Name).WithError(err).Warn("failed to create import event for function")
			continue
		}

		s.pipeline.AddMessage(event)
	}

	return nil
}
