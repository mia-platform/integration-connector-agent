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

package s3

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws/clients/s3"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"

	"github.com/sirupsen/logrus"
)

type S3 struct {
	logger *logrus.Logger
	client s3.Client
}

func New(logger *logrus.Logger, client s3.Client) *S3 {
	return &S3{
		logger: logger,
		client: client,
	}
}

func (s *S3) GetData(ctx context.Context, event *awssqsevents.CloudTrailEvent) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)

	bucketName, err := event.ResourceName()
	if err != nil {
		s.logger.WithError(err).Error("bucketName not found in request parameters")
		return nil, fmt.Errorf("%w: %s", commons.ErrInvalidEvent, err.Error())
	}

	tags, err := s.client.GetTags(ctx, bucketName)
	if err != nil {
		s.logger.WithError(err).Warn("failed to get S3 bucket tags")
		tags = make(commons.Tags)
	}

	relationships := []string{"account/" + event.Account}

	return json.Marshal(
		commons.
			NewAsset(bucketName, event.Detail.EventSource, commons.AWSAssetProvider).
			WithLocation(event.Detail.AWSRegion).
			WithTags(tags).
			WithRelationships(relationships).
			WithRawData(data),
	)
}
