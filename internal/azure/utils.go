// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package azure

import (
	"regexp"
	"strings"

	"github.com/mia-platform/integration-connector-agent/entities"
)

type EventRecord string

const (
	ManagedClusterEventSource           = "microsoft.containerservice/managedclusters"
	CognitiveServicesAccountEventSource = "microsoft.cognitiveservices/accounts"
	ContainerAppEventSource             = "microsoft.app/containerapps"
	FlexibleServerEventSource           = "microsoft.dbforpostgresql/flexibleservers"
	StorageAccountEventSource           = "microsoft.storage/storageaccounts"
	ResourceGroupEventSource            = "microsoft.resources/resourcegroups"
	ComputeVirtualMachineEventSource    = "microsoft.compute/virtualmachines"
	VirtualNetworkEventSource           = "microsoft.network/virtualnetworks"
	WebSitesEventSource                 = "microsoft.web/sites"

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
		ManagedClusterEventSource,
		CognitiveServicesAccountEventSource,
		ContainerAppEventSource,
		FlexibleServerEventSource,
		StorageAccountEventSource,
		ResourceGroupEventSource,
		ComputeVirtualMachineEventSource,
		VirtualNetworkEventSource,
		WebSitesEventSource,
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
