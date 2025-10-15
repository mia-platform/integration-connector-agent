// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package s3

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	aws "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/awsclient"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"

	"github.com/sirupsen/logrus"
)

type S3 struct {
	logger *logrus.Logger
	client aws.AWS
}

func New(logger *logrus.Logger, client aws.AWS) *S3 {
	return &S3{
		logger: logger,
		client: client,
	}
}

func (s *S3) GetData(ctx context.Context, event awssqsevents.IEvent) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)

	bucketName, err := event.ResourceName()
	if err != nil {
		s.logger.WithError(err).Error("bucketName not found in request parameters")
		return nil, fmt.Errorf("%w: %s", commons.ErrInvalidEvent, err.Error())
	}

	tags, err := s.client.GetBucketTags(ctx, bucketName)
	if err != nil {
		s.logger.WithError(err).Warn("failed to get S3 bucket tags")
		tags = make(commons.Tags)
	}

	relationships := []string{event.AccountID()}

	return json.Marshal(
		commons.
			NewAsset(bucketName, event.EventSource(), commons.AWSAssetProvider).
			WithLocation(event.GetRegion()).
			WithTags(tags).
			WithRelationships(relationships).
			WithRawData(data),
	)
}
