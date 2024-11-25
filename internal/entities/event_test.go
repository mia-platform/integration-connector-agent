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

package entities

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvent(t *testing.T) {
	e := &Event{
		ID:            "test",
		OperationType: Write,
		OriginalRaw:   []byte(`{"test": "test"}`),
	}

	eventCloned := e.Clone()

	require.Implements(t, (*PipelineEvent)(nil), e)
	require.Equal(t, "test", e.GetID())
	require.Equal(t, []byte(`{"test": "test"}`), e.Data())
	require.Equal(t, Write, e.Operation())
	e.WithData([]byte(`{"test": "test2"}`))
	require.Equal(t, []byte(`{"test": "test2"}`), e.Data())
	parsed, err := e.JSON()
	require.Equal(t, map[string]any{"test": "test2"}, parsed)
	require.NoError(t, err)
	require.Equal(t, &Event{
		ID:            "test",
		OperationType: Write,
		OriginalRaw:   []byte(`{"test": "test"}`),
	}, eventCloned)
	cloneParsed, err := e.JSON()
	require.Equal(t, map[string]any{"test": "test2"}, cloneParsed)
	require.NoError(t, err)
}
