// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package testutils

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func RandomString(tb testing.TB, n int) string {
	tb.Helper()

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		require.NoError(tb, err)
		b[i] = letters[num.Int64()]
	}
	return string(b)
}
