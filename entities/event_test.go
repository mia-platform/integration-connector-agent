// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package entities

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvent(t *testing.T) {
	e := &Event{
		PrimaryKeys:   PkFields{{Key: "test", Value: "test"}},
		OperationType: Write,
		OriginalRaw:   []byte(`{"test": "test"}`),
	}

	eventCloned := e.Clone()

	require.Implements(t, (*PipelineEvent)(nil), e)
	require.Equal(t, e.PrimaryKeys, e.GetPrimaryKeys())
	require.JSONEq(t, `{"test": "test"}`, string(e.Data()))
	require.Equal(t, Write, e.Operation())
	e.WithData([]byte(`{"test": "test2"}`))
	require.JSONEq(t, `{"test": "test2"}`, string(e.Data()))
	parsed, err := e.JSON()
	require.Equal(t, map[string]any{"test": "test2"}, parsed)
	require.NoError(t, err)
	require.Equal(t, &Event{
		PrimaryKeys:   PkFields{{Key: "test", Value: "test"}},
		OperationType: Write,
		OriginalRaw:   []byte(`{"test": "test"}`),
	}, eventCloned)
	cloneParsed, err := e.JSON()
	require.Equal(t, map[string]any{"test": "test2"}, cloneParsed)
	require.NoError(t, err)
}

func TestPkField(t *testing.T) {
	t.Run("isEmpty", func(t *testing.T) {
		pk := PkFields{}
		require.True(t, pk.IsEmpty())

		pk = PkFields{{Key: "test", Value: "test"}}
		require.False(t, pk.IsEmpty())
	})

	t.Run("map", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			pk := PkFields{{Key: "test", Value: "test"}, {Key: "test2", Value: "test2"}}
			require.Equal(t, map[string]string{
				"test":  "test",
				"test2": "test2",
			}, pk.Map())
		})

		t.Run("empty", func(t *testing.T) {
			pk := PkFields{}
			require.Equal(t, map[string]string{}, pk.Map())
		})

		t.Run("duplicated keys", func(t *testing.T) {
			pk := PkFields{
				{Key: "test", Value: "test"},
				{Key: "test", Value: "test2"},
			}
			require.Equal(t, map[string]string{
				"test": "test2",
			}, pk.Map())
		})
	})
}
