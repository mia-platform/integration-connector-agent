// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gcpclient

import (
	"context"
	"sync"

	"cloud.google.com/go/asset/apiv1/assetpb"
)

var _mockinterfaceimpltest GCP = &MockPubSub{} //nolint: unused

type MockPubSub struct {
	ListenError       error
	ListenAssert      func(ctx context.Context, handler ListenerFunc)
	listenInvoked     bool
	listenInvokedLock sync.Mutex

	CloseError       error
	closeInvoked     bool
	closrInvokedLock sync.Mutex

	ListAssetsResult      []*assetpb.Asset
	ListAssetsError       error
	listAssetsInvoked     bool
	listAssetsInvokedLock sync.Mutex

	ListBucketsResult      []*Bucket
	ListBucketsError       error
	listBucketsInvoked     bool
	listBucketsInvokedLock sync.Mutex

	ListFunctionsResult      []*Function
	ListFunctionsError       error
	listFunctionsInvoked     bool
	listFunctionsInvokedLock sync.Mutex
}

func (m *MockPubSub) Listen(ctx context.Context, handler ListenerFunc) error {
	m.listenInvokedLock.Lock()
	m.listenInvoked = true
	m.listenInvokedLock.Unlock()
	if m.ListenAssert != nil {
		m.ListenAssert(ctx, handler)
	}

	<-ctx.Done()
	return m.ListenError
}

func (m *MockPubSub) ListenInvoked() bool {
	m.listenInvokedLock.Lock()
	defer m.listenInvokedLock.Unlock()
	return m.listenInvoked
}

func (m *MockPubSub) Close() error {
	m.closrInvokedLock.Lock()
	defer m.closrInvokedLock.Unlock()
	m.closeInvoked = true
	return m.CloseError
}

func (m *MockPubSub) CloseInvoked() bool {
	m.closrInvokedLock.Lock()
	defer m.closrInvokedLock.Unlock()
	return m.closeInvoked
}

func (m *MockPubSub) ListAssets(_ context.Context) ([]*assetpb.Asset, error) {
	m.listAssetsInvokedLock.Lock()
	defer m.listAssetsInvokedLock.Unlock()

	m.listAssetsInvoked = true
	return m.ListAssetsResult, m.ListAssetsError
}

func (m *MockPubSub) ListAssetsInvoked() bool {
	m.listAssetsInvokedLock.Lock()
	defer m.listAssetsInvokedLock.Unlock()
	return m.listAssetsInvoked
}

func (m *MockPubSub) ListBuckets(_ context.Context) ([]*Bucket, error) {
	m.listBucketsInvokedLock.Lock()
	defer m.listBucketsInvokedLock.Unlock()

	m.listBucketsInvoked = true
	return m.ListBucketsResult, m.ListBucketsError
}

func (m *MockPubSub) ListBucketsInvoked() bool {
	m.listBucketsInvokedLock.Lock()
	defer m.listBucketsInvokedLock.Unlock()
	return m.listBucketsInvoked
}

func (m *MockPubSub) ListFunctions(_ context.Context) ([]*Function, error) {
	m.listFunctionsInvokedLock.Lock()
	defer m.listFunctionsInvokedLock.Unlock()

	m.listFunctionsInvoked = true
	return m.ListFunctionsResult, m.ListFunctionsError
}

func (m *MockPubSub) ListFunctionsInvoked() bool {
	m.listFunctionsInvokedLock.Lock()
	defer m.listFunctionsInvokedLock.Unlock()
	return m.listFunctionsInvoked
}
