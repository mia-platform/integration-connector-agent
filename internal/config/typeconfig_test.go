// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

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
