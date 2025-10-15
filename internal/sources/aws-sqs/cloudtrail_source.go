// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package awssqs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources"
	"github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/awsclient"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook/hmac"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type CloudTrailSourceConfig struct {
	QueueURL        string              `json:"queueUrl"`
	Region          string              `json:"region"`
	AccessKeyID     string              `json:"accessKeyId,omitempty"`
	SecretAccessKey config.SecretSource `json:"secretAccessKey,omitempty"`
	SessionToken    config.SecretSource `json:"sessionToken,omitempty"`

	WebhookPath    string              `json:"webhookPath,omitempty"`
	Authentication hmac.Authentication `json:"authentication,omitempty"`
}

func (c *CloudTrailSourceConfig) Validate() error {
	if c.QueueURL == "" {
		return errors.New("queueId must be provided")
	}

	return c.Authentication.Validate()
}

type CloudTrailSource struct {
	ctx      context.Context
	log      *logrus.Logger
	config   *CloudTrailSourceConfig
	pipeline pipeline.IPipelineGroup

	aws    awsclient.AWS
	sqs    *sqsConsumer
	router *swagger.Router[fiber.Handler, fiber.Router]
}

func NewCloudTrailSource(
	ctx context.Context,
	log *logrus.Logger,
	cfg config.GenericConfig,
	pipeline pipeline.IPipelineGroup,
	oasRouter *swagger.Router[fiber.Handler, fiber.Router],
) (sources.CloseableSource, error) {
	config, err := config.GetConfig[*CloudTrailSourceConfig](cfg)
	if err != nil {
		return nil, err
	}

	client, err := awsclient.New(ctx, log, awsclient.Config{
		QueueURL:        config.QueueURL,
		Region:          config.Region,
		AccessKeyID:     config.AccessKeyID,
		SecretAccessKey: config.SecretAccessKey.String(),
		SessionToken:    config.SessionToken.String(),
	})
	if err != nil {
		return nil, err
	}

	s := newCloudTrailSource(
		ctx,
		log,
		config,
		pipeline,
		oasRouter,
	)
	if err := s.init(client); err != nil {
		return nil, fmt.Errorf("failed to initialize inventory source: %w", err)
	}
	return s, nil
}

func newCloudTrailSource(
	ctx context.Context,
	log *logrus.Logger,
	config *CloudTrailSourceConfig,
	pipeline pipeline.IPipelineGroup,
	oasRouter *swagger.Router[fiber.Handler, fiber.Router],
) *CloudTrailSource {
	return &CloudTrailSource{
		ctx:      ctx,
		log:      log,
		config:   config,
		pipeline: pipeline,
		router:   oasRouter,
	}
}

func (s *CloudTrailSource) init(client awsclient.AWS) error {
	s.pipeline.Start(s.ctx)

	s.aws = client

	eventBuilder := awssqsevents.NewCloudTrailEventBuilder[*awssqsevents.CloudTrailEvent]()
	s.sqs = newSQS(s.ctx, s.log, s.pipeline, eventBuilder, s.aws)

	if s.config.WebhookPath != "" {
		s.log.WithField("webhookPath", s.config.WebhookPath).Info("Registering import webhook")
		if err := s.registerImportWebhook(); err != nil {
			return fmt.Errorf("failed to register import webhook: %w", err)
		}
	}
	return nil
}

func (s *CloudTrailSource) Close() error {
	if s.aws != nil {
		return s.aws.Close()
	}
	if s.sqs != nil {
		return s.sqs.Close()
	}
	return nil
}

func (s *CloudTrailSource) registerImportWebhook() error {
	apiPath := s.config.WebhookPath

	_, err := s.router.AddRoute(http.MethodPost, apiPath, s.webhookHandler, swagger.Definitions{})
	return err
}

func (s *CloudTrailSource) webhookHandler(c *fiber.Ctx) error {
	if err := s.config.Authentication.CheckSignature(c); err != nil {
		s.log.WithError(err).Error("error validating webhook request")
		return c.Status(http.StatusBadRequest).JSON(utils.ValidationError(err.Error()))
	}

	eventBuilder := awssqsevents.NewCloudTrailEventBuilder[*awssqsevents.CloudTrailImportEvent]()

	buckets, err := s.aws.ListBuckets(c.UserContext())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to list buckets: " + err.Error()))
	}
	for _, bucket := range buckets {
		importEvent := awssqsevents.CloudTrailImportEvent{
			Name:    bucket.Name,
			Source:  awssqsevents.CloudTrailEventStorageType,
			Region:  bucket.Region,
			Account: bucket.AccountID,
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

	functions, err := s.aws.ListFunctions(c.UserContext())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.InternalServerError("failed to list functions: " + err.Error()))
	}
	for _, function := range functions {
		importEvent := awssqsevents.CloudTrailImportEvent{
			Name:    function.Name,
			Source:  awssqsevents.CloudTrailEventFunctionType,
			Region:  function.Region,
			Account: function.AccountID,
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

	c.Status(http.StatusNoContent)
	return nil
}
