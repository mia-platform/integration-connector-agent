// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package awsclient

import (
	"context"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
)

const (
	S3EventSource     = "s3.amazonaws.com"
	LambdaEventSource = "lambda.amazonaws.com"
)

type ListenerFunc func(ctx context.Context, data []byte) error

type AWS interface {
	GetBucketTags(ctx context.Context, bucketName string) (commons.Tags, error)
	ListBuckets(ctx context.Context) ([]*Bucket, error)
	GetFunction(ctx context.Context, functionName string) (*Function, error)
	ListFunctions(ctx context.Context) ([]*Function, error)
	Listen(ctx context.Context, handler ListenerFunc) error
	Close() error
}

type Bucket struct {
	Name      string
	AccountID string
	Region    string
}

type Function struct {
	Name      string
	AccountID string
	Region    string
	ARN       string
	Tags      commons.Tags
}
