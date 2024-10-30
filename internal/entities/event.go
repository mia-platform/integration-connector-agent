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

import "github.com/tidwall/gjson"

type Operation int

const (
	Write Operation = iota
	Delete
)

type PipelineEvent interface {
	GetID() string
	RawData() []byte
	Type() Operation
	ParsedData() gjson.Result
}

type Event struct {
	ID            string
	OperationType Operation

	Raw    []byte
	Parsed gjson.Result
}

func (e Event) GetID() string {
	return e.ID
}

func (e Event) RawData() []byte {
	return e.Raw
}

func (e Event) ParsedData() gjson.Result {
	return e.Parsed
}

func (e Event) Type() Operation {
	return e.OperationType
}
