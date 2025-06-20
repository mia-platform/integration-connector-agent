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

package processors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/processors/filter"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/stretchr/testify/require"
)

type mockProcessor struct {
	processFunc func(data entities.PipelineEvent) (entities.PipelineEvent, error)
}

func (m *mockProcessor) Process(data entities.PipelineEvent) (entities.PipelineEvent, error) {
	return m.processFunc(data)
}

func TestProcessors_Process(t *testing.T) {
	tests := map[string]struct {
		name        string
		processors  []entities.Processor
		input       entities.PipelineEvent
		expected    entities.PipelineEvent
		expectedErr string
	}{
		"successful processing": {
			processors: []entities.Processor{
				&mockProcessor{
					processFunc: func(event entities.PipelineEvent) (entities.PipelineEvent, error) {
						event.WithData(append(event.Data(), []byte(" processed")...))
						return event, nil
					},
				},
			},
			input:    &entities.Event{OriginalRaw: []byte("test")},
			expected: &entities.Event{OriginalRaw: []byte("test processed")},
		},
		"processor error": {
			processors: []entities.Processor{
				&mockProcessor{
					processFunc: func(_ entities.PipelineEvent) (entities.PipelineEvent, error) {
						return nil, errors.New("processing error")
					},
				},
			},
			input:       &entities.Event{OriginalRaw: []byte("test")},
			expectedErr: "processing error",
		},
		"successful filter": {
			processors: []entities.Processor{
				&mockProcessor{
					processFunc: func(event entities.PipelineEvent) (entities.PipelineEvent, error) {
						return event, fmt.Errorf("%w: event filtered", filter.ErrEventToFilter)
					},
				},
				&mockProcessor{
					processFunc: func(_ entities.PipelineEvent) (entities.PipelineEvent, error) {
						panic("the event should be filtered")
					},
				},
			},
			input:       &entities.Event{OriginalRaw: []byte("test")},
			expectedErr: fmt.Sprintf("%s: event filtered", filter.ErrEventToFilter),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processors{processors: tt.processors}
			got, err := p.Process(t.Context(), tt.input)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected.Data(), got.Data())
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := map[string]struct {
		cfg config.Processors

		expectedErr string
	}{
		"unsupported processor": {
			cfg: config.Processors{
				{Type: "unsupported"},
			},

			expectedErr: ErrProcessorNotSupported.Error(),
		},
		"mapper processor": {
			cfg: config.Processors{
				{Type: Mapper, Raw: []byte(`{"type":"mapper","outputEvent":{"field":"{{ foo.bar }}"}}`)},
			},
		},
		"mapper processor - wrong config": {
			cfg: config.Processors{
				{Type: Mapper, Raw: []byte(`{"type":"mapper","outputEvent":invalid-json}`)},
			},
			expectedErr: "invalid character 'i' looking for beginning of value",
		},
		"filter processor": {
			cfg: config.Processors{
				{Type: Filter, Raw: []byte(`{"type":"filter","celExpression":"eventType == 'my-event-type'"}`)},
			},
		},
		"filter processor - wrong config": {
			cfg: config.Processors{
				{Type: Filter, Raw: []byte(`{"type":"filter","celExpression":"foo"}`)},
			},
			expectedErr: "ERROR: <input>:1:1: undeclared reference to 'foo' (in container '')\n | foo\n | ^",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			log, _ := test.NewNullLogger()

			proc, err := New(log, tt.cfg)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, proc.processors, len(tt.cfg))
		})
	}
}
