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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/internal"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	pg := &pipeline.Group{}
	log, _ := test.NewNullLogger()

	options := &ConsumerOptions{
		Ctx: t.Context(),
		Log: log,
	}

	t.Run("fails on invalid config", func(t *testing.T) {
		_, err := New(options, config.GenericConfig{
			Type: "gcppubsub",
			Raw:  []byte(`{"projectId": "", "topicName": "", "subscriptionId": ""}`),
		}, pg, &eventBuilderMock{})

		require.ErrorIs(t, err, config.ErrConfigNotValid)
	})

	t.Run("succeeds with valid config", func(t *testing.T) {
		consumer, err := New(options, config.GenericConfig{
			Type: "gcppubsub",
			Raw:  []byte(`{"projectId": "test-project", "topicName": "test-topic", "subscriptionId": "test-subscription"}`),
		}, pg, &eventBuilderMock{})

		require.NoError(t, err)
		require.NotNil(t, consumer)
	})
}

func TestClientIntegrationWithEventBuilder(t *testing.T) {
	log, _ := test.NewNullLogger()

	// This test uses a cancelable context to simulate the real-world scenario
	// when using the underlying gcp sdk `subscription.Receive` method.
	// Which blocks until the context is canceled or an error is received.
	// https://pkg.go.dev/cloud.google.com/go/pubsub@v1.49.0#Subscription.Receive
	t.Run("client lifecycle with cancelable context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		o := &ConsumerOptions{Ctx: ctx, Log: log}
		pg := &pipelineGroupMock{}
		e := &eventBuilderMock{}
		config := &Config{}
		client := &internal.MockPubSub{}

		consumer, err := newWithClient(o, pg, e, config, client)
		require.NoError(t, err)
		require.NotNil(t, consumer)

		require.True(t, pg.startInvoked)

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
		o := &ConsumerOptions{Ctx: ctx, Log: log}
		pg := &pipelineGroupMock{
			assertAddMessage: func(data entities.PipelineEvent) {
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
		config := &Config{}

		client := &internal.MockPubSub{
			ListenAssert: func(ctx context.Context, handler internal.ListenerFunc) {
				require.NotNil(t, ctx)
				require.NotNil(t, handler)

				// Simulate receiving a message from Pub/Sub
				err := handler(ctx, dataFromPubSub)
				require.NoError(t, err)
			},
		}

		consumer, err := newWithClient(o, pg, e, config, client)
		require.NoError(t, err)
		require.NotNil(t, consumer)

		require.True(t, pg.startInvoked)

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
		o := &ConsumerOptions{Ctx: ctx, Log: log}
		pg := &pipelineGroupMock{
			assertAddMessage: func(data entities.PipelineEvent) {
				require.NotNil(t, data)
				require.Equal(t, "some-type", data.GetType())
			},
		}
		e := &eventBuilderMock{
			assertData: func(data []byte) {
				require.Equal(t, dataFromPubSub, data)
			},
			returnedErr: fmt.Errorf("some error from event builder"),
		}
		config := &Config{}

		client := &internal.MockPubSub{
			ListenAssert: func(ctx context.Context, handler internal.ListenerFunc) {
				require.NotNil(t, ctx)
				require.NotNil(t, handler)

				// Simulate receiving a message from Pub/Sub
				err := handler(ctx, dataFromPubSub)
				require.Error(t, err, "some error from event builder")
			},
		}

		consumer, err := newWithClient(o, pg, e, config, client)
		require.NoError(t, err)
		require.NotNil(t, consumer)

		require.True(t, pg.startInvoked)

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
		o := &ConsumerOptions{Ctx: ctx, Log: log}
		pg := &pipelineGroupMock{
			assertAddMessage: func(data entities.PipelineEvent) {
				require.NotNil(t, data)
				require.Equal(t, "some-type", data.GetType())
			},
		}
		e := &eventBuilderMock{
			GetPipelineEventFunc: func(ctx context.Context, data []byte) (entities.PipelineEvent, error) {
				require.NotNil(t, ctx)
				require.NotNil(t, data)
				if string(data) == "failing payload" {
					return nil, fmt.Errorf("failed to process payload")
				}

				require.Equal(t, dataFromPubSub, data)
				return &entities.Event{
					Type: "some-type",
				}, nil
			},
		}
		config := &Config{}

		var handlerRef internal.ListenerFunc
		var handlerRefLock sync.Mutex
		client := &internal.MockPubSub{
			ListenAssert: func(ctx context.Context, handler internal.ListenerFunc) {
				require.NotNil(t, ctx)
				require.NotNil(t, handler)

				// Simulate receiving a message from Pub/Sub
				handlerRefLock.Lock()
				handlerRef = handler
				handlerRefLock.Unlock()
			},
		}

		consumer, err := newWithClient(o, pg, e, config, client)
		require.NoError(t, err)
		require.NotNil(t, consumer)

		require.True(t, pg.startInvoked)

		// Allow some time for the goroutine to start
		time.Sleep(10 * time.Millisecond)
		require.True(t, client.ListenInvoked())
		require.False(t, client.CloseInvoked())

		handlerRefLock.Lock()
		defer handlerRefLock.Unlock()
		require.NotNil(t, handlerRef)

		err = handlerRef(ctx, dataFromPubSub)
		require.NoError(t, err)

		err = handlerRef(ctx, []byte("failing payload"))
		require.Error(t, err, "failing to process payload")

		err = handlerRef(ctx, dataFromPubSub)
		require.NoError(t, err)

		cancel()
		time.Sleep(10 * time.Millisecond)
		require.True(t, client.CloseInvoked())
	})
}
