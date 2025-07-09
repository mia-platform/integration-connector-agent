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

package gcppubsub

import (
	"context"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/internal"

	"github.com/sirupsen/logrus"
)

type pubsubConsumer struct {
	config   *pubSubConfig
	pipeline pipeline.IPipelineGroup
	log      *logrus.Logger
	client   internal.PubSub
}

type pubSubConfig struct {
	log                *logrus.Logger
	ctx                context.Context
	ProjectID          string
	TopicName          string
	SubscriptionID     string
	AckDeadlineSeconds int
	CredentialsJSON    string
}

func newPubSub(cfg *pubSubConfig, pipeline pipeline.IPipelineGroup, eventBuilder entities.EventBuilder) (*pubsubConsumer, error) {
	client, err := internal.New(
		cfg.ctx,
		cfg.log,
		internal.PubSubConfig{
			ProjectID:          cfg.ProjectID,
			TopicName:          cfg.TopicName,
			SubscriptionID:     cfg.SubscriptionID,
			AckDeadlineSeconds: cfg.AckDeadlineSeconds,
			CredentialsJSON:    cfg.CredentialsJSON,
		},
	)
	if err != nil {
		return nil, err
	}

	return newPubSubWithClient(pipeline, eventBuilder, cfg, client)
}

func newPubSubWithClient(pipeline pipeline.IPipelineGroup, eventBuilder entities.EventBuilder, config *pubSubConfig, client internal.PubSub) (*pubsubConsumer, error) {
	pipeline.Start(config.ctx)

	go func(ctx context.Context, log *logrus.Logger, client internal.PubSub) {
		err := client.Listen(ctx, func(ctx context.Context, data []byte) error {
			event, err := eventBuilder.GetPipelineEvent(ctx, data)
			if err != nil {
				return err
			}

			pipeline.AddMessage(event)
			return nil
		})
		if err != nil {
			log.WithError(err).Error("error listening to GCP Pub/Sub")
		}

		client.Close()
	}(config.ctx, config.log, client)

	return &pubsubConsumer{
		log:      config.log,
		config:   config,
		client:   client,
		pipeline: pipeline,
	}, nil
}

func (g *pubsubConsumer) Close() error {
	return g.client.Close()
}
