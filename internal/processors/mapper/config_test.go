// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package mapper

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("unmarshal json", func(t *testing.T) {
		cfg := Config{}
		err := json.Unmarshal([]byte(`{"outputEvent":{"foo": "{{ .bar.taz }}"}}`), &cfg)
		require.NoError(t, err)
		require.Equal(t, Config{
			OutputEvent: []byte(`{"foo": "{{ .bar.taz }}"}`),
		}, cfg)
	})

	t.Run("validate", func(t *testing.T) {
		cfg := Config{}
		require.NoError(t, cfg.Validate())
	})
}
