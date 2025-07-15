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
