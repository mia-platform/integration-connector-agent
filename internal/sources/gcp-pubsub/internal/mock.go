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

package internal

import (
	"context"
	"sync"
)

type MockPubSub struct {
	ListenError       error
	ListenAssert      func(ctx context.Context, handler ListenerFunc)
	listenInvoked     bool
	listenInvokedLock sync.Mutex

	CloseError      error
	closeInvoked    bool
	closInvokedLock sync.Mutex
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
	m.closInvokedLock.Lock()
	defer m.closInvokedLock.Unlock()
	m.closeInvoked = true
	return m.CloseError
}

func (m *MockPubSub) CloseInvoked() bool {
	m.closInvokedLock.Lock()
	defer m.closInvokedLock.Unlock()
	return m.closeInvoked
}
