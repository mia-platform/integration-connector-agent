// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

func NewGCPRunServiceDataAdapter(client runservice.Client) commons.DataAdapter[gcppubsubevents.IInventoryEvent] {
	return &GCPRunServiceDataAdapter{
		client: client,
	}
}

func (g *GCPRunServiceDataAdapter) GetData(ctx context.Context, event gcppubsubevents.IInventoryEvent) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)

	runServiceName := strings.TrimPrefix(event.Name(), "//run.googleapis.com/")

	service, err := g.client.GetService(ctx, runServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCP run service: %w", err)
	}

	name, location := nameAndLocationFromRunName(runServiceName)

	return json.Marshal(
		commons.NewAsset(name, event.AssetType(), commons.GCPAssetProvider).
			WithLocation(location).
			WithTags(service.Labels).
			WithRelationships(event.Ancestors()).
			WithRawData(data),
	)
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
