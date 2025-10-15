// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gcpclient

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBucketAssetName(t *testing.T) {
	bucket := &Bucket{Name: "test-bucket"}
	require.Equal(t, "//storage.googleapis.com/test-bucket", bucket.AssetName())
}
