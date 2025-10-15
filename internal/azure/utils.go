// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package azure

import (
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

	CognitiveServicesAccountEventSource    = "microsoft.cognitiveservices/accounts"
	CognitiveServicesDeploymentEventSource = "microsoft.cognitiveservices/accounts/deployments"
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
				relationships = append(relationships, "subscription/"+group)
			case "resourceGroupName":
				relationships = append(relationships, "resourceGroup/"+group)
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
		CognitiveServicesAccountEventSource,
		CognitiveServicesDeploymentEventSource,
	}

	eventSource := strings.ToLower(event.OperationName)
	for _, source := range allSources {
		switch eventSource {
		case source + "/delete":
			return source
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
