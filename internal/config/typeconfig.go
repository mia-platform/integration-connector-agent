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

package config

import (
	"encoding/json"
	"fmt"
)

type GenericConfig struct {
	Type string `json:"type"`

	Raw []byte `json:"-"`
}

func (w *GenericConfig) UnmarshalJSON(data []byte) error {
	var writerConfig struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &writerConfig); err != nil {
		return err
	}

	w.Type = writerConfig.Type
	w.Raw = data

	return nil
}

type Validator interface {
	Validate() error
}

func GetConfig[T Validator](config GenericConfig) (T, error) {
	var cfg T
	if err := json.Unmarshal(config.Raw, &cfg); err != nil {
		return cfg, err
	}

	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("%w: %w", ErrConfigNotValid, err)
	}

	return cfg, nil
}
