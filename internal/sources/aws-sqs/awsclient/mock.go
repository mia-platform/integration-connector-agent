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
	"sync"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
)

type AWSMock struct {
	ListenError       error
	ListenAssert      func(ctx context.Context, handler ListenerFunc)
	listenInvoked     bool
	listenInvokedLock sync.Mutex

	GetBucketTagsResult commons.Tags
	GetBucketTagsError  error

	ListBucketsResult      []*Bucket
	ListBucketsError       error
	listBucketsInvoked     bool
	listBucketsInvokedLock sync.Mutex

	GetFunctionResult *Function
	GetFunctionError  error

	ListFunctionsResult      []*Function
	ListFunctionsError       error
	listFunctionsInvoked     bool
	listFunctionsInvokedLock sync.Mutex

	CloseError      error
	closeInvoked    bool
	closInvokedLock sync.Mutex
}

func (m *AWSMock) Listen(ctx context.Context, handler ListenerFunc) error {
	m.listenInvokedLock.Lock()
	m.listenInvoked = true
	m.listenInvokedLock.Unlock()
	if m.ListenAssert != nil {
		m.ListenAssert(ctx, handler)
	}

	<-ctx.Done()
	return m.ListenError
}

func (m *AWSMock) ListenInvoked() bool {
	m.listenInvokedLock.Lock()
	defer m.listenInvokedLock.Unlock()
	return m.listenInvoked
}

func (m *AWSMock) Close() error {
	m.closInvokedLock.Lock()
	defer m.closInvokedLock.Unlock()
	m.closeInvoked = true
	return m.CloseError
}

func (m *AWSMock) CloseInvoked() bool {
	m.closInvokedLock.Lock()
	defer m.closInvokedLock.Unlock()
	return m.closeInvoked
}

func (m *AWSMock) ListBuckets(_ context.Context) ([]*Bucket, error) {
	m.listBucketsInvokedLock.Lock()
	defer m.listBucketsInvokedLock.Unlock()

	m.listBucketsInvoked = true
	return m.ListBucketsResult, m.ListBucketsError
}

func (m *AWSMock) ListBucketsInvoked() bool {
	m.listBucketsInvokedLock.Lock()
	defer m.listBucketsInvokedLock.Unlock()
	return m.listBucketsInvoked
}

func (m *AWSMock) ListFunctions(_ context.Context) ([]*Function, error) {
	m.listFunctionsInvokedLock.Lock()
	defer m.listFunctionsInvokedLock.Unlock()

	m.listFunctionsInvoked = true
	return m.ListFunctionsResult, m.ListFunctionsError
}

func (m *AWSMock) ListFunctionsInvoked() bool {
	m.listFunctionsInvokedLock.Lock()
	defer m.listFunctionsInvokedLock.Unlock()
	return m.listFunctionsInvoked
}

func (m *AWSMock) GetBucketTags(ctx context.Context, bucketName string) (commons.Tags, error) {
	return m.GetBucketTagsResult, m.GetBucketTagsError
}

func (m *AWSMock) GetFunction(ctx context.Context, functionName string) (*Function, error) {
	return m.GetFunctionResult, m.GetFunctionError
}
