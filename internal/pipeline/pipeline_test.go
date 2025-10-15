// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/processors"
	fakesink "github.com/mia-platform/integration-connector-agent/internal/sinks/fake"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipeline(t *testing.T) {
	log, _ := test.NewNullLogger()
	model := &fakesink.Config{}
	proc := &processors.Processors{}

	t.Run("message pipelines is correctly managed adding messages", func(t *testing.T) {
		w := fakesink.New(model)
		p, err := New(log, proc, w)
		require.NoError(t, err)

		runPipeline(t, p)

		id := "fake event"
		testCases := map[string]struct {
			event *entities.Event

			expectedOperation entities.Operation
		}{
			"default operation": {
				event: &entities.Event{
					PrimaryKeys:   entities.PkFields{{Key: "key", Value: id}},
					OperationType: entities.Write,
				},
				expectedOperation: entities.Write,
			},
			"write operation": {
				event: &entities.Event{
					PrimaryKeys:   entities.PkFields{{Key: "key", Value: id}},
					OperationType: entities.Write,
				},
				expectedOperation: entities.Write,
			},
			"delete operation": {
				event: &entities.Event{
					PrimaryKeys:   entities.PkFields{{Key: "key", Value: id}},
					OperationType: entities.Delete,
				},
				expectedOperation: entities.Delete,
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				tc.event.OriginalRaw = []byte(`{}`)

				p.AddMessage(tc.event)

				assert.Eventually(t, func() bool {
					return len(w.Calls()) == 1
				}, 1*time.Second, 10*time.Millisecond)
				data := &entities.Event{
					PrimaryKeys:   entities.PkFields{{Key: "key", Value: id}},
					OperationType: tc.expectedOperation,
					OriginalRaw:   []byte(`{}`),
				}
				assert.Equal(t, fakesink.Call{
					Operation: tc.expectedOperation,
					Data:      data,
				}, w.Calls().LastCall())
				w.ResetCalls()
			})
		}
	})

	t.Run("on channel closed, the pipeline stops", func(t *testing.T) {
		w := fakesink.New(model)
		p, err := New(log, proc, w)
		require.NoError(t, err)

		go func(t *testing.T) {
			t.Helper()
			time.Sleep(10 * time.Millisecond)

			if pipeline, ok := p.(*Pipeline); ok {
				eventChannel := pipeline.eventChan
				close(eventChannel)
			}
		}(t)

		err = p.Start(t.Context())
		assert.NoError(t, err)
	})

	t.Run("on context done, close channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		w := fakesink.New(model)
		p, err := New(log, proc, w)
		require.NoError(t, err)

		go func(t *testing.T, cancel context.CancelFunc) {
			t.Helper()
			time.Sleep(10 * time.Millisecond)
			cancel()
		}(t, cancel)

		err = p.Start(ctx)
		assert.EqualError(t, err, "context canceled")
	})

	t.Run("on sink error, the pipeline skips the element and logs - write", func(t *testing.T) {
		w := fakesink.New(&fakesink.Config{
			Mocks: []fakesink.Mock{
				{Error: errors.New("fake error")},
			},
		})

		log, hook := test.NewNullLogger()

		p, err := New(log, proc, w)
		require.NoError(t, err)

		runPipeline(t, p)

		id := "fake event"
		p.AddMessage(&entities.Event{
			PrimaryKeys:   entities.PkFields{{Key: "key", Value: id}},
			OperationType: entities.Write,
			OriginalRaw:   []byte(`{}`),
		})

		assert.Eventually(t, func() bool {
			return len(w.Calls()) == 1
		}, 1*time.Second, 10*time.Millisecond)
		event := &entities.Event{
			PrimaryKeys:   entities.PkFields{{Key: "key", Value: id}},
			OperationType: entities.Write,
			OriginalRaw:   []byte(`{}`),
		}
		assert.Equal(t, fakesink.Call{
			Operation: entities.Write,
			Data:      event,
		}, w.Calls().LastCall())

		assert.Equal(t, "error writing data to sink", hook.LastEntry().Message)
	})

	t.Run("on error, the pipeline skips the element and logs - delete", func(t *testing.T) {
		w := fakesink.New(&fakesink.Config{
			Mocks: []fakesink.Mock{
				{Error: errors.New("fake error")},
			},
		})

		log, hook := test.NewNullLogger()

		p, err := New(log, proc, w)
		require.NoError(t, err)

		runPipeline(t, p)

		id := "fake event"
		p.AddMessage(&entities.Event{
			PrimaryKeys:   entities.PkFields{{Key: "key", Value: id}},
			OperationType: entities.Delete,
		})

		assert.Eventually(t, func() bool {
			return len(w.Calls()) == 1
		}, 1*time.Second, 10*time.Millisecond)
		assert.Equal(t, fakesink.Call{
			Operation: entities.Delete,
			Data: &entities.Event{
				PrimaryKeys:   entities.PkFields{{Key: "key", Value: id}},
				OperationType: entities.Delete,
			},
		}, w.Calls().LastCall())
		assert.Equal(t, "error writing data to sink", hook.LastEntry().Message)
	})

	t.Run("filter event when filter returns false", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		w := fakesink.New(model)
		proc, err := processors.New(log, config.Processors{
			{
				Type: processors.Filter,
				Raw:  []byte(`{"type":"filter","celExpression":"false"}`),
			},
		})
		require.NoError(t, err)

		p, err := New(log, proc, w)
		require.NoError(t, err)
		runPipeline(t, p)

		p.AddMessage(&entities.Event{
			PrimaryKeys:   entities.PkFields{{Key: "key", Value: "fake event"}},
			Type:          "event-type",
			OperationType: entities.Write,

			OriginalRaw: []byte(`{"type":"event-type"}`),
		})

		assert.Eventually(t, func() bool {
			return len(w.Calls()) < 1
		}, 1*time.Second, 100*time.Millisecond)

		assert.Empty(t, w.Calls())
		logErrorProcessingDataMessageCount := 0
		for _, entry := range hook.AllEntries() {
			if entry.Message == "error processing data" {
				logErrorProcessingDataMessageCount++
			}
		}

		assert.Equal(t, 0, logErrorProcessingDataMessageCount)
	})
}

func runPipeline(t *testing.T, p IPipeline) {
	t.Helper()

	go func(t *testing.T) {
		t.Helper()

		err := p.Start(t.Context())
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("error starting pipeline: %v", err)
		}
	}(t)
}
