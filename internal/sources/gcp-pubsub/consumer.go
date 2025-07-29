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
	"github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/gcpclient"

	"github.com/sirupsen/logrus"
)

type pubsubConsumer struct {
	pipeline pipeline.IPipelineGroup
	log      *logrus.Logger
	client   gcpclient.GCP
}

func newPubSub(
	ctx context.Context,
	log *logrus.Logger,
	pipeline pipeline.IPipelineGroup,
	eventBuilder entities.EventBuilder,
	client gcpclient.GCP,
) *pubsubConsumer {
	go func(ctx context.Context, log *logrus.Logger, client gcpclient.GCP) {
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
	}(ctx, log, client)

	return &pubsubConsumer{
		log:      log,
		client:   client,
		pipeline: pipeline,
	}
}

func (g *pubsubConsumer) Close() error {
	return g.client.Close()
}
