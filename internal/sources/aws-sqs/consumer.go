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

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/internal"

	"github.com/sirupsen/logrus"
)

type AWSConsumer struct {
	config   *Config
	pipeline pipeline.IPipelineGroup
	log      *logrus.Logger
	client   internal.SQS
}

type ConsumerOptions struct {
	Ctx context.Context
	Log *logrus.Logger
}

func New(options *ConsumerOptions, cfg config.GenericConfig, pipeline pipeline.IPipelineGroup, eventBuilder entities.EventBuilder) (*AWSConsumer, error) {
	config, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return nil, err
	}

	client, err := internal.New(options.Ctx, options.Log, internal.Config{
		QueueURL:        config.QueueURL,
		Region:          config.Region,
		AccessKeyID:     config.AccessKeyID,
		SecretAccessKey: config.SecretAccessKey,
		SessionToken:    config.SessionToken,
	})
	if err != nil {
		return nil, err
	}

	return newWithClient(options, pipeline, eventBuilder, config, client)
}

func newWithClient(options *ConsumerOptions, pipeline pipeline.IPipelineGroup, eventBuilder entities.EventBuilder, config *Config, client internal.SQS) (*AWSConsumer, error) {
	pipeline.Start(options.Ctx)

	go func(ctx context.Context, log *logrus.Logger, client internal.SQS) {
		err := client.Listen(ctx, func(ctx context.Context, data []byte) error {
			event, err := eventBuilder.GetPipelineEvent(ctx, data)
			if err != nil {
				return err
			}

			log.WithFields(logrus.Fields{
				"queueUrl":         config.QueueURL,
				"eventPrimaryKeys": event.GetPrimaryKeys(),
			}).Debug("received event from AWS SQS queue")

			pipeline.AddMessage(event)
			return nil
		})
		if err != nil {
			log.WithField("queueUrl", config.QueueURL).WithError(err).Error("error listening to AWS SQS queue")
		}

		client.Close()
	}(options.Ctx, options.Log, client)

	return &AWSConsumer{
		config:   config,
		pipeline: pipeline,
		log:      options.Log,
		client:   client,
	}, nil
}

func (a *AWSConsumer) Close() error {
	return a.client.Close()
}
