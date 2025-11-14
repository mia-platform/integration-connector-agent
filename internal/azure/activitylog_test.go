// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package azure

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/stretchr/testify/assert"
)

func TestEntityOperationType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		record       ActivityLogEventRecord
		expectedType entities.Operation
	}{
		"Delete operation": {
			record: ActivityLogEventRecord{
				OperationName: "Microsoft.Compute/virtualMachines/delete",
			},
			expectedType: entities.Delete,
		},
		"Delete action operation": {
			record: ActivityLogEventRecord{
				OperationName: "Microsoft.Compute/virtualMachines/delete/action",
			},
			expectedType: entities.Delete,
		},
		"Write operation": {
			record: ActivityLogEventRecord{
				OperationName: "Microsoft.Compute/virtualMachines/write",
			},
			expectedType: entities.Write,
		},
		"List keys operation": {
			record: ActivityLogEventRecord{
				OperationName: "Microsoft.KeyVault/vaults/listkeys/action",
			},
			expectedType: entities.Write,
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.expectedType, test.record.entityOperationType())
		})
	}
}
