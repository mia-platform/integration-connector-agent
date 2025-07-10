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

package gcpclient

import (
	"context"
	"sync"
)

type MockPubSub struct {
	ListenError       error
	ListenAssert      func(ctx context.Context, handler ListenerFunc)
	listenInvoked     bool
	listenInvokedLock sync.Mutex

	CloseError       error
	closeInvoked     bool
	closrInvokedLock sync.Mutex

	ListBucketsResult      []*Bucket
	ListBucketsError       error
	listBucketsInvoked     bool
	listBucketsInvokedLock sync.Mutex
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

func (m *MockPubSub) ListBuckets(ctx context.Context) ([]*Bucket, error) {
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
