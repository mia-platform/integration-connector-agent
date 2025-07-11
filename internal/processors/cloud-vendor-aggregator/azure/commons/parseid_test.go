// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
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

package commons

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelationshipFromID(t *testing.T) {
	testCases := map[string]struct {
		id                    string
		expectedRelationships []string
	}{
		"valid ID with subscription and resource group": {
			id:                    "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myResourceGroup/providers/Microsoft.Web/site/myfunction",
			expectedRelationships: []string{"subscription/12345678-1234-1234-1234-123456789012", "resourceGroup/myResourceGroup"},
		},
		"valid ID with subscription and lowercase resource group": {
			id:                    "/subscriptions/12345678-1234-1234-1234-123456789012/resourcegroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/myVM",
			expectedRelationships: []string{"subscription/12345678-1234-1234-1234-123456789012", "resourceGroup/myResourceGroup"},
		},
		"invalid ID with subscription and resource group": {
			id:                    "subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myResourceGroup/providers/Microsoft.Storage/storageAccounts/myStorageAccount",
			expectedRelationships: []string{},
		},
		"invalid ID": {
			id:                    "not/a/valid/id",
			expectedRelationships: []string{},
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			parsedRelationships := RelationshipFromID(test.id)
			assert.Equal(t, test.expectedRelationships, parsedRelationships)
		})
	}
}
