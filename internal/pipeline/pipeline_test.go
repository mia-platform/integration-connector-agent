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

package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/internal/entities"
	fakewriter "github.com/mia-platform/integration-connector-agent/internal/sinks/fake"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestPipeline(t *testing.T) {
	log := logrus.New()
	model := &fakewriter.Config{
		OutputModel: map[string]any{},
	}

	t.Run("message pipelines is correctly managed adding messages", func(t *testing.T) {
		w := fakewriter.New(model)
		p, err := NewPipeline(logrus.NewEntry(log), w)
		require.NoError(t, err)

		runPipeline(t, p)

		id := "fake event"
		testCases := map[string]struct {
			event *entities.Event

			expectedOperation entities.Operation
		}{
			"default operation": {
				event: &entities.Event{
					ID:            id,
					OperationType: entities.Write,
				},
				expectedOperation: entities.Write,
			},
			"write operation": {
				event: &entities.Event{
					ID:            id,
					OperationType: entities.Write,
				},
				expectedOperation: entities.Write,
			},
			"delete operation": {
				event: &entities.Event{
					ID:            id,
					OperationType: entities.Delete,
				},
				expectedOperation: entities.Delete,
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				tc.event.OriginalRaw = []byte(`{}`)

				p.AddMessage(tc.event)

				require.Eventually(t, func() bool {
					data := &entities.Event{
						ID:            id,
						OperationType: tc.expectedOperation,
						OriginalRaw:   []byte(`{}`),
					}
					if tc.expectedOperation == entities.Write {
						data.WithData(map[string]any{})
					}

					require.Equal(t, fakewriter.Call{
						Operation: tc.expectedOperation,
						Data:      data,
					}, w.Calls().LastCall())
					return true
				}, 1*time.Second, 10*time.Millisecond)
			})
		}
	})

	t.Run("on channel closed, the pipeline stops", func(t *testing.T) {
		ctx := context.Background()
		w := fakewriter.New(model)
		p, err := NewPipeline(logrus.NewEntry(log), w)
		require.NoError(t, err)

		go func(t *testing.T) {
			t.Helper()
			time.Sleep(10 * time.Millisecond)

			eventChannel := getPipeline(t, p).eventChan
			close(eventChannel)
		}(t)

		err = p.Start(ctx)
		require.NoError(t, err)
	})

	t.Run("on context done, close channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		w := fakewriter.New(model)
		p, err := NewPipeline(logrus.NewEntry(log), w)
		require.NoError(t, err)

		go func(t *testing.T, cancel context.CancelFunc) {
			t.Helper()
			time.Sleep(10 * time.Millisecond)
			cancel()
		}(t, cancel)

		err = p.Start(ctx)
		require.EqualError(t, err, "context canceled")
	})

	t.Run("on error, the pipeline skips the element and logs - write", func(t *testing.T) {
		w := fakewriter.New(model)
		w.AddMock(fakewriter.Mock{
			Error: errors.New("fake error"),
		})

		log, hook := test.NewNullLogger()

		p, err := NewPipeline(logrus.NewEntry(log), w)
		require.NoError(t, err)

		runPipeline(t, p)

		id := "fake event"
		p.AddMessage(&entities.Event{
			ID:            id,
			OperationType: entities.Write,
			OriginalRaw:   []byte(`{}`),
		})

		require.Eventually(t, func() bool {
			event := &entities.Event{
				ID:            id,
				OperationType: entities.Write,
				OriginalRaw:   []byte(`{}`),
			}
			event.WithData(map[string]any{})
			require.Equal(t, fakewriter.Call{
				Operation: entities.Write,
				Data:      event,
			}, w.Calls().LastCall())
			return true
		}, 1*time.Second, 10*time.Millisecond)
		require.Equal(t, "error writing data", hook.LastEntry().Message)
	})

	t.Run("on error, the pipeline skips the element and logs - delete", func(t *testing.T) {
		w := fakewriter.New(model)
		w.AddMock(fakewriter.Mock{
			Error: errors.New("fake error"),
		})

		log, hook := test.NewNullLogger()

		p, err := NewPipeline(logrus.NewEntry(log), w)
		require.NoError(t, err)

		runPipeline(t, p)

		id := "fake event"
		p.AddMessage(&entities.Event{
			ID:            id,
			OperationType: entities.Delete,
		})

		require.Eventually(t, func() bool {
			require.Equal(t, fakewriter.Call{
				Operation: entities.Delete,
				Data: &entities.Event{
					ID:            id,
					OperationType: entities.Delete,
				},
			}, w.Calls().LastCall())
			return true
		}, 1*time.Second, 10*time.Millisecond)
		require.Equal(t, "error deleting data", hook.LastEntry().Message)
	})
}

func getPipeline[T entities.PipelineEvent](t *testing.T, p IPipeline[T]) *Pipeline[T] {
	t.Helper()

	pipeline, ok := p.(*Pipeline[T])
	require.True(t, ok)

	return pipeline
}

func runPipeline(t *testing.T, p IPipeline[entities.PipelineEvent]) {
	t.Helper()

	ctx := context.Background()

	go func(t *testing.T) {
		t.Helper()

		err := p.Start(ctx)
		require.NoError(t, err)
	}(t)
}
