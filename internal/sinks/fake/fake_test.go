// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package fakewriter

import (
	"errors"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"

	"github.com/stretchr/testify/require"
)

func TestImplementWriter(t *testing.T) {
	config := &Config{}

	t.Run("implement writer", func(t *testing.T) {
		require.Implements(t, (*sinks.Sink[entities.PipelineEvent])(nil), New(config))
	})

	t.Run("stub write", func(t *testing.T) {
		f := New(config)

		event := &entities.Event{
			PrimaryKeys:   entities.PkFields{{Key: "key", Value: "id"}},
			OperationType: entities.Write,
		}
		err := f.WriteData(t.Context(), event)
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
			PrimaryKeys:   entities.PkFields{{Key: "key", Value: "id"}},
			OperationType: entities.Delete,
		}
		err := f.WriteData(t.Context(), event)
		require.NoError(t, err)

		require.Len(t, f.Calls(), 1)
		require.Equal(t, Call{
			Data:      event,
			Operation: entities.Delete,
		}, f.Calls().LastCall())
	})

	t.Run("ResetCalls clean calls", func(t *testing.T) {
		f := New(config)

		event := &entities.Event{
			PrimaryKeys:   entities.PkFields{{Key: "key", Value: "id"}},
			OperationType: entities.Write,
		}
		err := f.WriteData(t.Context(), event)
		require.NoError(t, err)

		require.Len(t, f.Calls(), 1)
		f.ResetCalls()
		require.Empty(t, f.Calls())
	})

	t.Run("mock error write", func(t *testing.T) {
		f := New(config)

		event := &entities.Event{
			PrimaryKeys:   entities.PkFields{{Key: "key", Value: "id"}},
			OperationType: entities.Write,
		}
		f.AddMock(Mock{
			Error: errors.New("mock error"),
		})
		err := f.WriteData(t.Context(), event)
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
			PrimaryKeys:   entities.PkFields{{Key: "key", Value: "id"}},
			OperationType: entities.Delete,
		}
		f.AddMock(Mock{
			Error: errors.New("mock error"),
		})
		err := f.WriteData(t.Context(), event)
		require.EqualError(t, err, "mock error")

		require.Len(t, f.Calls(), 1)
		require.Equal(t, Call{
			Data:      event,
			Operation: entities.Delete,
		}, f.Calls().LastCall())
	})
}
