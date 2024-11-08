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

package fakewriter

import (
	"context"
	"errors"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/writer"

	"github.com/stretchr/testify/require"
)

func TestImplementWriter(t *testing.T) {
	config := &Config{
		OutputModel: map[string]any{},
	}

	t.Run("implement writer", func(t *testing.T) {
		require.Implements(t, (*writer.Writer[entities.PipelineEvent])(nil), New(config))
	})

	t.Run("stub write", func(t *testing.T) {
		f := New(config)

		event := &entities.Event{
			ID: "id",
		}
		err := f.Write(context.Background(), event)
		require.NoError(t, err)

		require.Len(t, f.Calls(), 1)
		require.Equal(t, Call{
			Data:      event,
			Operation: entities.Write,
		}, f.Calls().LastCall())
	})

	t.Run("stub delete", func(t *testing.T) {
		f := New(config)

		event := &entities.Event{
			ID: "id",
		}
		err := f.Delete(context.Background(), event)
		require.NoError(t, err)

		require.Len(t, f.Calls(), 1)
		require.Equal(t, Call{
			Data:      event,
			Operation: entities.Delete,
		}, f.Calls().LastCall())
	})

	t.Run("mock error write", func(t *testing.T) {
		f := New(config)

		event := &entities.Event{
			ID: "id",
		}
		f.AddMock(Mock{
			Error: errors.New("mock error"),
		})
		err := f.Write(context.Background(), event)
		require.EqualError(t, err, "mock error")

		require.Len(t, f.Calls(), 1)
		require.Equal(t, Call{
			Data:      event,
			Operation: entities.Write,
		}, f.Calls().LastCall())
	})

	t.Run("mock error delete", func(t *testing.T) {
		f := New(config)

		event := &entities.Event{
			ID: "id",
		}
		f.AddMock(Mock{
			Error: errors.New("mock error"),
		})
		err := f.Delete(context.Background(), event)
		require.EqualError(t, err, "mock error")

		require.Len(t, f.Calls(), 1)
		require.Equal(t, Call{
			Data:      event,
			Operation: entities.Delete,
		}, f.Calls().LastCall())
	})

	t.Run("output model", func(t *testing.T) {
		f := New(config)

		require.Equal(t, config.OutputModel, f.OutputModel())
	})
}
