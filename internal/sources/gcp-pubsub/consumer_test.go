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

package gcppubsub

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/gcpclient"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestClientIntegrationWithEventBuilder(t *testing.T) {
	log, _ := test.NewNullLogger()

	// This test uses a cancelable context to simulate the real-world scenario
	// when using the underlying gcp sdk `subscription.Receive` method.
	// Which blocks until the context is canceled or an error is received.
	// https://pkg.go.dev/cloud.google.com/go/pubsub@v1.49.0#Subscription.Receive
	t.Run("client lifecycle with cancelable context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		pg := &pipeline.PipelineGroupMock{}
		e := &eventBuilderMock{}
		client := &gcpclient.MockPubSub{}

		consumer := newPubSub(ctx, log, pg, e, client)
		require.NotNil(t, consumer)

		// Allow some time for the goroutine to start
		time.Sleep(10 * time.Millisecond)
		require.True(t, client.ListenInvoked())
		require.False(t, client.CloseInvoked())

		cancel()

		// Allow some time for the goroutine to run the close-up logic
		time.Sleep(10 * time.Millisecond)

		require.True(t, client.CloseInvoked())
	})

	t.Run("messages are correctly sent to the pipeline", func(t *testing.T) {
		dataFromPubSub := []byte("test-data-from-pubsub")

		ctx, cancel := context.WithCancel(t.Context())
		pg := &pipeline.PipelineGroupMock{
			AssertAddMessage: func(data entities.PipelineEvent) {
				require.NotNil(t, data)
				require.Equal(t, "some-type", data.GetType())
			},
		}
		e := &eventBuilderMock{
			assertData: func(data []byte) {
				require.Equal(t, dataFromPubSub, data)
			},
			returnedEvent: &entities.Event{
				Type: "some-type",
			},
		}

		client := &gcpclient.MockPubSub{
			ListenAssert: func(ctx context.Context, handler gcpclient.ListenerFunc) {
				require.NotNil(t, ctx)
				require.NotNil(t, handler)

				// Simulate receiving a message from Pub/Sub
				err := handler(ctx, dataFromPubSub)
				require.NoError(t, err)
			},
		}

		consumer := newPubSub(ctx, log, pg, e, client)
		require.NotNil(t, consumer)

		// Allow some time for the goroutine to start
		time.Sleep(10 * time.Millisecond)
		require.True(t, client.ListenInvoked())
		require.False(t, client.CloseInvoked())

		cancel()
		time.Sleep(10 * time.Millisecond)
		require.True(t, client.CloseInvoked())
	})

	t.Run("messages are not sent to the pipeline on event builder error", func(t *testing.T) {
		dataFromPubSub := []byte("test-data-from-pubsub")

		ctx, cancel := context.WithCancel(t.Context())
		pg := &pipeline.PipelineGroupMock{
			AssertAddMessage: func(data entities.PipelineEvent) {
				require.NotNil(t, data)
				require.Equal(t, "some-type", data.GetType())
			},
		}
		e := &eventBuilderMock{
			assertData: func(data []byte) {
				require.Equal(t, dataFromPubSub, data)
			},
			returnedErr: errors.New("some error from event builder"),
		}

		client := &gcpclient.MockPubSub{
			ListenAssert: func(ctx context.Context, handler gcpclient.ListenerFunc) {
				require.NotNil(t, ctx)
				require.NotNil(t, handler)

				// Simulate receiving a message from Pub/Sub
				err := handler(ctx, dataFromPubSub)
				require.Error(t, err, "some error from event builder")
			},
		}

		consumer := newPubSub(ctx, log, pg, e, client)
		require.NotNil(t, consumer)

		// Allow some time for the goroutine to start
		time.Sleep(10 * time.Millisecond)
		require.True(t, client.ListenInvoked())
		require.False(t, client.CloseInvoked())

		cancel()
		time.Sleep(10 * time.Millisecond)
		require.True(t, client.CloseInvoked())
	})

	t.Run("can process multiple messages", func(t *testing.T) {
		dataFromPubSub := []byte("test-data-from-pubsub")

		ctx, cancel := context.WithCancel(t.Context())

		pg := &pipeline.PipelineGroupMock{
			AssertAddMessage: func(data entities.PipelineEvent) {
				require.NotNil(t, data)
				require.Equal(t, "some-type", data.GetType())
			},
		}
		e := &eventBuilderMock{
			GetPipelineEventFunc: func(ctx context.Context, data []byte) (entities.PipelineEvent, error) {
				require.NotNil(t, ctx)
				require.NotNil(t, data)
				if string(data) == "failing payload" {
					return nil, errors.New("failed to process payload")
				}

				require.Equal(t, dataFromPubSub, data)
				return &entities.Event{
					Type: "some-type",
				}, nil
			},
		}

		var handlerRef gcpclient.ListenerFunc
		var handlerRefLock sync.Mutex
		client := &gcpclient.MockPubSub{
			ListenAssert: func(ctx context.Context, handler gcpclient.ListenerFunc) {
				require.NotNil(t, ctx)
				require.NotNil(t, handler)

				// Simulate receiving a message from Pub/Sub
				handlerRefLock.Lock()
				handlerRef = handler
				handlerRefLock.Unlock()
			},
		}

		consumer := newPubSub(ctx, log, pg, e, client)
		require.NotNil(t, consumer)

		// Allow some time for the goroutine to start
		time.Sleep(10 * time.Millisecond)
		require.True(t, client.ListenInvoked())
		require.False(t, client.CloseInvoked())

		handlerRefLock.Lock()
		defer handlerRefLock.Unlock()
		require.NotNil(t, handlerRef)

		require.NoError(t, handlerRef(ctx, dataFromPubSub))
		require.Error(t, handlerRef(ctx, []byte("failing payload")), "failing to process payload")
		require.NoError(t, handlerRef(ctx, dataFromPubSub))

		cancel()
		time.Sleep(10 * time.Millisecond)
		require.True(t, client.CloseInvoked())
	})
}
