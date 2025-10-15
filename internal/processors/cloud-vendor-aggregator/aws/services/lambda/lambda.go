// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package lambda

import (
	"context"
	"encoding/json"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/awsclient"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"

	"github.com/sirupsen/logrus"
)

type Lambda struct {
	logger *logrus.Logger
	client awsclient.AWS
}

func New(logger *logrus.Logger, client awsclient.AWS) *Lambda {
	return &Lambda{
		logger: logger,
		client: client,
	}
}

func (l *Lambda) GetData(ctx context.Context, event awssqsevents.IEvent) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)

	name := l.lambdaName(event)
	if name == "" {
		l.logger.Error("functionName not found in request parameters or response elements")
		return nil, commons.ErrInvalidEvent
	}

	functionDetails, err := l.client.GetFunction(ctx, name)
	if err != nil {
		l.logger.WithError(err).Warn("failed to get Lambda function details")
	}

	tags := make(commons.Tags)
	if functionDetails != nil {
		tags = functionDetails.Tags
	}

	relationships := []string{event.AccountID()}

	return json.Marshal(
		commons.NewAsset(name, event.EventSource(), commons.AWSAssetProvider).
			WithLocation(event.GetRegion()).
			WithTags(tags).
			WithRelationships(relationships).
			WithRawData(data),
	)
}

func (l *Lambda) lambdaName(event awssqsevents.IEvent) string {
	name, err := event.ResourceName()
	if err == nil {
		return name
	}
	l.logger.WithError(err).Debug("failed to get resource name from event")
	return ""
}
