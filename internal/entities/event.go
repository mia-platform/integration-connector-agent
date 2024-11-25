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

import "encoding/json"

type Operation int

const (
	Write Operation = iota
	Delete
)

type PipelineEvent interface {
	GetID() string
	GetType() string

	Data() []byte
	Operation() Operation
	WithData([]byte)
	JSON() (map[string]any, error)
	Clone() PipelineEvent
}

type Event struct {
	ID            string
	Type          string
	OperationType Operation

	OriginalRaw []byte
	jsonData    map[string]any
}

func (e Event) GetID() string {
	return e.ID
}

func (e Event) Data() []byte {
	return e.OriginalRaw
}

func (e Event) Operation() Operation {
	return e.OperationType
}

func (e Event) JSON() (map[string]any, error) {
	if e.jsonData != nil {
		return e.jsonData, nil
	}
	parsed := map[string]any{}

	if err := json.Unmarshal(e.OriginalRaw, &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

func (e *Event) WithData(raw []byte) {
	e.OriginalRaw = raw
}

func (e *Event) Clone() PipelineEvent {
	return &Event{
		ID:            e.ID,
		OperationType: e.OperationType,
		OriginalRaw:   e.OriginalRaw,
		Type:          e.Type,

		jsonData: e.jsonData,
	}
}

func (e Event) GetType() string {
	return e.Type
}
