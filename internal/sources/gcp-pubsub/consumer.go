// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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
