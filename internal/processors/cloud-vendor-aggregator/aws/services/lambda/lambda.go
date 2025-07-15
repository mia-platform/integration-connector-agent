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
