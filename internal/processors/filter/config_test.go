// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package filter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	cfg := Config{}

	require.NoError(t, cfg.Validate())
}
