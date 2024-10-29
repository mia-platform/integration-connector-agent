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

package aggregator

import (
	"github.com/mia-platform/data-connector-agent/internal/mapper"
	"github.com/mia-platform/data-connector-agent/internal/writer"
)

type Pipeline[T writer.DataWithIdentifier] struct {
	writer writer.Writer[T]
	mapper mapper.Mapper
}

func NewPipeline[T writer.DataWithIdentifier](writer writer.Writer[T], mapper mapper.Mapper) *Pipeline[T] {
	return &Pipeline[T]{writer: writer, mapper: mapper}
}
