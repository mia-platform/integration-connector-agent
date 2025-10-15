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
