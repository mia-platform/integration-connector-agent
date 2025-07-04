package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/azure/client"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	azureactivitylogeventhubevents "github.com/mia-platform/integration-connector-agent/internal/sources/azure-activity-log-event-hub/events"
)

const (
	EventSource = "microsoft.storage/storageaccounts"
)

type AzureStorage struct {
	client client.Client
}

func New(getter client.Client) *AzureStorage {
	return &AzureStorage{
		client: getter,
	}
}

func (a *AzureStorage) GetData(context context.Context, event *azureactivitylogeventhubevents.ActivityLogEventRecord) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)
	entity, found := event.Properties["entity"]
	if !found {
		return nil, fmt.Errorf("entity not found in event properties")
	}

	resource, err := a.client.GetByID(entity.(string))
	if err != nil {
		return nil, fmt.Errorf("failed to get resource by ID: %w", err)
	}

	tags := make(commons.Tags)
	for key, value := range resource.Tags {
		if value != nil {
			tags[key] = *value
		}
	}

	asset := &commons.Asset{
		Name:          *resource.Name,
		Type:          *resource.Type,
		Provider:      commons.AzureAssetProvider,
		Location:      *resource.Location,
		Tags:          tags,
		Relationships: relationshipFromID(entity.(string)),
		RawData:       data,
	}

	return json.Marshal(asset)
}

func relationshipFromID(id string) []string {
	relationships := make([]string, 0)

	regex := regexp.MustCompile(`^/subscriptions/(?P<subscriptionId>[^/]+)/resourceGroups/(?P<resourceGroupName>[^/]+)/`)
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
