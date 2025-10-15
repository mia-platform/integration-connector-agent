// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package awssqs

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/awsclient"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestClientIntegrationWithEventBuilder(t *testing.T) {
	log, _ := test.NewNullLogger()

	t.Run("messages are correctly sent to the pipeline", func(t *testing.T) {
		dataFromPubSub := []byte("test-data-from-sqs")

		ctx, cancel := context.WithCancel(t.Context())
		pg := &awssqsevents.PipelineGroupMock{
			AssertAddMessage: func(data entities.PipelineEvent) {
				require.NotNil(t, data)
				require.Equal(t, "some-type", data.GetType())
			},
		}
		e := &awssqsevents.EventBuilderMock{
			AssertData: func(data []byte) {
				require.Equal(t, dataFromPubSub, data)
			},
			ReturnedEvent: &entities.Event{
				Type: "some-type",
			},
		}

		client := &awsclient.AWSMock{
			ListenAssert: func(ctx context.Context, handler awsclient.ListenerFunc) {
				require.NotNil(t, ctx)
				require.NotNil(t, handler)

				// Simulate receiving a message from Pub/Sub
				err := handler(ctx, dataFromPubSub)
				require.NoError(t, err)
			},
		}

		consumer := newSQS(ctx, log, pg, e, client)
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
		dataFromPubSub := []byte("test-data-from-sqs")

		ctx, cancel := context.WithCancel(t.Context())
		pg := &awssqsevents.PipelineGroupMock{
			AssertAddMessage: func(data entities.PipelineEvent) {
				require.NotNil(t, data)
				require.Equal(t, "some-type", data.GetType())
			},
		}
		e := &awssqsevents.EventBuilderMock{
			AssertData: func(data []byte) {
				require.Equal(t, dataFromPubSub, data)
			},
			ReturnedErr: errors.New("some error from event builder"),
		}

		client := &awsclient.AWSMock{
			ListenAssert: func(ctx context.Context, handler awsclient.ListenerFunc) {
				require.NotNil(t, ctx)
				require.NotNil(t, handler)

				// Simulate receiving a message from Pub/Sub
				err := handler(ctx, dataFromPubSub)
				require.Error(t, err, "some error from event builder")
			},
		}

		consumer := newSQS(ctx, log, pg, e, client)
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
		dataFromPubSub := []byte("test-data-from-sqs")

		ctx, cancel := context.WithCancel(t.Context())
		pg := &awssqsevents.PipelineGroupMock{
			AssertAddMessage: func(data entities.PipelineEvent) {
				require.NotNil(t, data)
				require.Equal(t, "some-type", data.GetType())
			},
		}
		e := awssqsevents.EventBuilderMock{
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

		var handlerRef awsclient.ListenerFunc
		var handlerRefLock sync.Mutex
		client := &awsclient.AWSMock{
			ListenAssert: func(ctx context.Context, handler awsclient.ListenerFunc) {
				require.NotNil(t, ctx)
				require.NotNil(t, handler)

				// Simulate receiving a message from Pub/Sub
				handlerRefLock.Lock()
				handlerRef = handler
				handlerRefLock.Unlock()
			},
		}

		consumer := newSQS(ctx, log, pg, e, client)
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
