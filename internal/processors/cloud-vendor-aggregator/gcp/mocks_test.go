// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

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
