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

package gcpaggregator

import (
	"context"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp/clients/runservice"
	storageclient "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp/clients/storage"
)

type MockFnService struct {
	GetServiceAssert func(ctx context.Context, name string)
	GetServiceResult *runservice.Service
	GetServiceErr    error
}

func (m *MockFnService) GetService(ctx context.Context, name string) (*runservice.Service, error) {
	if m.GetServiceAssert != nil {
		m.GetServiceAssert(ctx, name)
	}
	return m.GetServiceResult, m.GetServiceErr
}
func (m *MockFnService) Close() error {
	return nil
}

type MockStorageClient struct {
	GetBucketAssert func(ctx context.Context, name string)
	GetBucketResult *storageclient.Bucket
	GetBucketErr    error
}

func (m *MockStorageClient) GetBucket(ctx context.Context, name string) (*storageclient.Bucket, error) {
	if m.GetBucketAssert != nil {
		m.GetBucketAssert(ctx, name)
	}
	return m.GetBucketResult, m.GetBucketErr
}
func (m *MockStorageClient) Close() error {
	return nil
}
