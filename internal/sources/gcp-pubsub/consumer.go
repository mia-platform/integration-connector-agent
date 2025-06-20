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

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/internal"

	"github.com/sirupsen/logrus"
)

type ConsumerOptions struct {
	Ctx context.Context
	Log *logrus.Logger
}

type GCPConsumer struct {
	config   *Config
	pipeline pipeline.IPipelineGroup
	log      *logrus.Logger
	client   internal.PubSub
}

func New(options *ConsumerOptions, cfg config.GenericConfig, pipeline pipeline.IPipelineGroup, eventBuilder EventBuilder) (*GCPConsumer, error) {
	config, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return nil, err
	}

	client, err := internal.New(options.Ctx, options.Log, internal.PubSubConfig{
		ProjectID:          config.ProjectID,
		TopicName:          config.TopicName,
		SubscriptionID:     config.SubscriptionID,
		AckDeadlineSeconds: config.AckDeadlineSeconds,
	})
	if err != nil {
		return nil, err
	}

	return newWithClient(options, pipeline, eventBuilder, config, client)
}

func newWithClient(options *ConsumerOptions, pipeline pipeline.IPipelineGroup, eventBuilder EventBuilder, config *Config, client internal.PubSub) (*GCPConsumer, error) {
	pipeline.Start(options.Ctx)

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
	}(options.Ctx, options.Log, client)

	return &GCPConsumer{
		log:      options.Log,
		config:   config,
		client:   client,
		pipeline: pipeline,
	}, nil
}

func (g *GCPConsumer) Close() error {
	g.client.Close()
	return nil
}
