// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package services

import (
	"context"

	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"
)

type AWSService interface {
	GetData(ctx context.Context, event *awssqsevents.CloudTrailEvent) ([]byte, error)
}
