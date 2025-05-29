// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
package sources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSourceTypes(t *testing.T) {
	require.Equal(t, "jira", Jira)
	require.Equal(t, "console", Console)
	require.Equal(t, "github", Github)
}
