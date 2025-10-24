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
