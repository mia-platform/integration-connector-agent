package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp/clients/runservice"
	gcppubsubevents "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/events"
)

const (
	RunServiceAssetType = "run.googleapis.com/Service"
)

type GCPRunServiceDataAdapter struct {
	client runservice.Client
}

func NewGCPRunServiceDataAdapter(ctx context.Context, client runservice.Client) commons.DataAdapter[*gcppubsubevents.InventoryEvent] {
	return &GCPRunServiceDataAdapter{
		client: client,
	}
}

func (g *GCPRunServiceDataAdapter) GetData(ctx context.Context, event *gcppubsubevents.InventoryEvent) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)

	// client, err := run.NewServicesClient(ctx, options)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create GCP run service client: %w", err)
	// }
	// defer client.Close()

	// runServiceName := strings.TrimPrefix(event.Asset.Name, "//run.googleapis.com/")
	// service, err := client.GetService(ctx, &runpb.GetServiceRequest{
	// 	Name: runServiceName,
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get GCP run service: %w", err)
	// }

	runServiceName := strings.TrimPrefix(event.Asset.Name, "//run.googleapis.com/")

	service, err := g.client.GetService(ctx, runServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCP run service: %w", err)
	}

	name, location := nameAndLocationFromRunName(runServiceName)
	asset := &commons.Asset{
		Name:          name,
		Type:          event.Asset.AssetType,
		Provider:      commons.GCPAssetProvider,
		Location:      location,
		Tags:          service.Labels,
		Relationships: event.Asset.Ancestors,
		RawData:       data,
	}

	return json.Marshal(asset)
}

func nameAndLocationFromRunName(runName string) (string, string) {
	regex := regexp.MustCompile(`projects/[^/]+/locations/(?P<location>[^/]+)/services/(?P<name>[^/]+)`)
	groupNames := regex.SubexpNames()
	var location, name string
	for _, match := range regex.FindAllStringSubmatch(runName, -1) {
		for groupIdx, groupMatch := range match {
			groupName := groupNames[groupIdx]
			switch groupName {
			case "location":
				location = groupMatch
			case "name":
				name = groupMatch
			}
		}
	}

	return name, location
}
