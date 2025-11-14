// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidationErrors(t *testing.T) {
	t.Run("returns correct validation error", func(t *testing.T) {
		err := ValidationError("test")
		require.Equal(t, &HTTPError{
			Message: "test",
			Error:   "Validation Error",
		}, err)
	})
}
