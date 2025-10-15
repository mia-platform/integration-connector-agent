// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenericConfig(t *testing.T) {
	t.Run("unmarshal json", func(t *testing.T) {
		data := []byte(`{"type":"test","key":"value"}`)

		config := &GenericConfig{}
		err := json.Unmarshal(data, config)
		require.NoError(t, err)
		require.Equal(t, &GenericConfig{
			Type: "test",
			Raw:  data,
		}, config)
	})
}

type mockTestGetConfig struct {
	Key string `json:"key"`
}

func (t mockTestGetConfig) Validate() error {
	return nil
}

func TestGetConfig(t *testing.T) {
	t.Run("get config", func(t *testing.T) {
		config := GenericConfig{
			Type: "test",
			Raw:  []byte(`{"type":"test","key":"value"}`),
		}

		cfg, err := GetConfig[mockTestGetConfig](config)
		require.NoError(t, err)
		require.Equal(t, mockTestGetConfig{
			Key: "value",
		}, cfg)
	})
}
