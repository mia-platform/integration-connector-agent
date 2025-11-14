// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package awssqs

import (
	"context"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/awsclient"

	"github.com/sirupsen/logrus"
)

type sqsConsumer struct {
	pipeline pipeline.IPipelineGroup
	log      *logrus.Logger
	client   awsclient.AWS
}

func newSQS(
	ctx context.Context,
	log *logrus.Logger,
	pipeline pipeline.IPipelineGroup,
	eventBuilder entities.EventBuilder,
	client awsclient.AWS,
) *sqsConsumer {
	go func(ctx context.Context, log *logrus.Logger, client awsclient.AWS) {
		err := client.Listen(ctx, func(ctx context.Context, data []byte) error {
			event, err := eventBuilder.GetPipelineEvent(ctx, data)
			if err != nil {
				return err
			}

			log.WithFields(logrus.Fields{
				"eventPrimaryKeys": event.GetPrimaryKeys(),
			}).Debug("received event from AWS SQS queue")

			pipeline.AddMessage(event)
			return nil
		})
		if err != nil {
			log.WithError(err).Error("error listening to AWS SQS queue")
		}

		client.Close()
	}(ctx, log, client)

	return &sqsConsumer{
		pipeline: pipeline,
		log:      log,
		client:   client,
	}
}

func (a *sqsConsumer) Close() error {
	return a.client.Close()
}
