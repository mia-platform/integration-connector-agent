// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package cloudvendoraggregator

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name          string
		cfg           config.Config
		expectedError error
	}{
		{
			name: "unsupported cloud vendor",
			cfg: config.Config{
				CloudVendorName: "unsupported",
			},
			expectedError: config.ErrInvalidCloudVendor,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := test.NewNullLogger()

			processor, err := New(logger, tc.cfg)
			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
				require.Nil(t, processor)
			} else {
				require.NoError(t, err)
				require.NotNil(t, processor)
			}
		})
	}
}
