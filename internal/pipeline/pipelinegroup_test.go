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
		sink1 := fakesink.New(&fakesink.Config{})
		sink2 := fakesink.New(&fakesink.Config{})

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
