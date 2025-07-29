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
