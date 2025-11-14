// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package pipeline

import (
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/processors"
	fakesink "github.com/mia-platform/integration-connector-agent/internal/sinks/fake"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestPipelineGroup(t *testing.T) {
	logger, _ := test.NewNullLogger()

	proc1, err := processors.New(logger, config.Processors{
		{
			Type: processors.Mapper,
			Raw:  []byte(`{"type":"mapper","outputEvent":{"field":"some"}}`),
		},
	})
	require.NoError(t, err)

	proc2, err := processors.New(logger, config.Processors{
		{
			Type: processors.Mapper,
			Raw:  []byte(`{"type":"mapper","outputEvent":{"field":"other"}}`),
		},
	})
	require.NoError(t, err)

	t.Run("multiple pipeline", func(t *testing.T) {
		sink1 := fakesink.New(&fakesink.Config{}, logger)
		sink2 := fakesink.New(&fakesink.Config{}, logger)

		p1, err := New(logger, proc1, sink1)
		require.NoError(t, err)
		p2, err := New(logger, proc2, sink2)
		require.NoError(t, err)

		pg := NewGroup(logger, p1, p2)

		pg.Start(t.Context())

		event := &entities.Event{
			PrimaryKeys: entities.PkFields{{Key: "id", Value: "123"}},
			OriginalRaw: []byte(`{"id":"123"}`),
		}
		pg.AddMessage(event)

		require.Eventually(t, func() bool {
			return len(sink1.Calls()) == 1 && len(sink2.Calls()) == 1
		}, time.Second, 10*time.Millisecond)

		require.JSONEq(t, `{"field":"some"}`, string(sink1.Calls().LastCall().Data.Data()))
		require.JSONEq(t, `{"field":"other"}`, string(sink2.Calls().LastCall().Data.Data()))
	})
}
