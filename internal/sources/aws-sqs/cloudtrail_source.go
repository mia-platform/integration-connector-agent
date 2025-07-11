// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package awssqs

import (
	"context"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"
	"github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/internal"

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
}

func (c *CloudTrailSourceConfig) Validate() error {
	if c.QueueURL == "" {
		return fmt.Errorf("queueId must be provided")
	}

	return nil
}

type CloudTrailSource struct {
	ctx      context.Context
	log      *logrus.Logger
	config   *CloudTrailSourceConfig
	pipeline pipeline.IPipelineGroup

	sqs    internal.SQS
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

	client, err := internal.New(ctx, log, internal.Config{
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

func (s *CloudTrailSource) init(client internal.SQS) error {
	s.pipeline.Start(s.ctx)

	s.sqs = client

	eventBuilder := awssqsevents.NewCloudTrailEventBuilder()
	newSQS(s.ctx, s.log, s.pipeline, eventBuilder, s.sqs)
	return nil
}

func (s *CloudTrailSource) Close() error {
	if s.sqs != nil {
		return s.sqs.Close()
	}
	return nil
}
