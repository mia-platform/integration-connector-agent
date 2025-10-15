// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package azure

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/stretchr/testify/assert"
)

func TestRelationshipFromID(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			parsedRelationships := RelationshipFromID(test.id)
			assert.Equal(t, test.expectedRelationships, parsedRelationships)
		})
	}
}

func TestEventIsForResourceType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		event        *ActivityLogEventRecord
		resourceType string
		expected     bool
	}{
		"event update tags for functions": {
			event: &ActivityLogEventRecord{
				OperationName: "MICROSOFT.RESOURCES/TAGS/WRITE",
				ResourceID:    "/SUBSCRIPTIONS/123/RESOURCEGROUPS/MYRESOURCEGROUP/PROVIDERS/MICROSOFT.WEB/SITES/MYFUNCTIONAPP",
			},
			resourceType: WebSitesEventSource,
			expected:     true,
		},
		"event for functions": {
			event: &ActivityLogEventRecord{
				OperationName: "MICROSOFT.WEB/SITES/WRITE",
				ResourceID:    "/SUBSCRIPTIONS/123/RESOURCEGROUPS/MYRESOURCEGROUP/PROVIDERS/MICROSOFT.WEB/SITES/MYFUNCTIONAPP",
			},
			resourceType: WebSitesEventSource,
			expected:     true,
		},
		"delete event for functions": {
			event: &ActivityLogEventRecord{
				OperationName: "MICROSOFT.WEB/SITES/DELETE",
				ResourceID:    "/SUBSCRIPTIONS/123/RESOURCEGROUPS/MYRESOURCEGROUP/PROVIDERS/MICROSOFT.WEB/SITES/MYFUNCTIONAPP",
			},
			resourceType: WebSitesEventSource,
			expected:     false,
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			result := EventIsForSource(test.event, test.resourceType)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestPrimaryKeys(t *testing.T) {
	t.Parallel()

	resourceID := "/SUBSCRIPTIONS/00000000-0000-0000-0000-000000000000/RESOURCEGROUPS/GROUP/PROVIDERS/MICROSOFT.COMPUTE/VIRTUALMACHINESCALESETS/SCALESET"

	expectedKeys := entities.PkFields{
		{
			Key:   "resourceId",
			Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/group/providers/microsoft.compute/virtualmachinescalesets/scaleset",
		},
	}

	assert.Equal(t, expectedKeys, primaryKeys(resourceID))
}
