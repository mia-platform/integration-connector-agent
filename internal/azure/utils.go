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

package azure

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mia-platform/integration-connector-agent/entities"
)

const (
	StorageAccountEventSource = "microsoft.storage/storageaccounts"

	WebSitesEventSource = "microsoft.web/sites"

	ComputeVirtualMachineEventSource = "microsoft.compute/virtualmachines"
	ComputeDiskEventSource           = "microsoft.compute/disks"

	VirtualNetworkEventSource         = "microsoft.network/virtualnetworks"
	NetworkInterfaceEventSource       = "microsoft.network/networkinterfaces"
	NetworkSecurityGroupEventSource   = "microsoft.network/networksecuritygroups"
	NetworkPublicIPAddressEventSource = "microsoft.network/publicipaddresses"

	TagsEventSource = "microsoft.resources/tags"
)

func RelationshipFromID(id string) []string {
	relationships := make([]string, 0)

	regex := regexp.MustCompile(`^/subscriptions/(?P<subscriptionId>[^/]+)/resource[gG]roups/(?P<resourceGroupName>[^/]+)/`)
	groupNames := regex.SubexpNames()
	for _, match := range regex.FindAllStringSubmatch(id, -1) {
		for groupIdx, group := range match {
			name := groupNames[groupIdx]
			switch name {
			case "subscriptionId":
				relationships = append(relationships, fmt.Sprintf("subscription/%s", group))
			case "resourceGroupName":
				relationships = append(relationships, fmt.Sprintf("resourceGroup/%s", group))
			}
		}
	}

	return relationships
}

func EventIsForSource(event *ActivityLogEventRecord, resourceType string) bool {
	eventSource := strings.ToLower(event.OperationName)
	resourceID := strings.ToLower(event.ResourceID)

	return eventSource == resourceType+"/write" ||
		(eventSource == TagsEventSource+"/write" && strings.Contains(resourceID, resourceType))
}

func EventSourceFromEvent(event *ActivityLogEventRecord) string {
	allSources := []string{
		StorageAccountEventSource,
		WebSitesEventSource,
		ComputeVirtualMachineEventSource,
		ComputeDiskEventSource,
		VirtualNetworkEventSource,
		NetworkInterfaceEventSource,
		NetworkSecurityGroupEventSource,
		NetworkPublicIPAddressEventSource,
	}

	eventSource := strings.ToLower(event.OperationName)
	for _, source := range allSources {
		switch eventSource {
		case source + "/write":
			return source
		case TagsEventSource + "/write":
			resourceID := strings.ToLower(event.ResourceID)
			if strings.Contains(resourceID, source) {
				return source
			}
		}
	}

	return ""
}

func primaryKeys(resourceID string) entities.PkFields {
	return entities.PkFields{
		{
			Key:   "resourceId",
			Value: strings.ToLower(resourceID),
		},
	}
}
