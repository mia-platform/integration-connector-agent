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

package mapper

import (
	"encoding/json"
	"strings"

	"github.com/tidwall/gjson"
)

type Mapper struct {
	operations []operation
}

func (m Mapper) Process(input []byte) ([]byte, error) {
	output := []byte("{}")
	var err error
	for _, operation := range m.operations {
		output, err = operation.apply(input, output)
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

func New(cfg Config) (*Mapper, error) {
	model, err := json.Marshal(cfg.OutputEvent)
	if err != nil {
		return nil, err
	}

	ops, err := generateOperations(gjson.ParseBytes(model))
	if err != nil {
		return nil, err
	}

	return &Mapper{
		operations: ops,
	}, nil
}

func generateOperations(jsonData gjson.Result) ([]operation, error) {
	result := []operation{}
	var resError error

	var walk func(data gjson.Result, keyPrefix string)
	walk = func(data gjson.Result, keyPrefix string) {
		data.ForEach(func(key, value gjson.Result) bool {
			keyToUpdate := key.String()
			if keyPrefix != "" {
				keyToUpdate = strings.Join([]string{keyPrefix, key.String()}, ".")
			}

			if value.IsObject() || value.IsArray() {
				walk(value, keyToUpdate)
				return true
			}

			if key.Exists() && key.String() != "" {
				operation, err := newOperation(keyToUpdate, value)
				if err != nil {
					resError = err
					return false
				}
				result = append(result, operation)
			}

			return true
		})
	}

	walk(jsonData, "")

	return result, resError
}
