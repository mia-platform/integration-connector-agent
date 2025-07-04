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

package gcppubsubevents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/entities"
)

var (
	ErrMalformedEvent = errors.New("malformed event")
)

type InventoryEventBuilder struct{}

func NewInventoryEventBuilder() entities.EventBuilder {
	return &InventoryEventBuilder{}
}

func (b *InventoryEventBuilder) GetPipelineEvent(_ context.Context, data []byte) (entities.PipelineEvent, error) {
	rawEvent := InventoryEvent{}
	if err := json.Unmarshal(data, &rawEvent); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMalformedEvent, err.Error())
	}

	return &entities.Event{
		PrimaryKeys:   b.primaryKeys(rawEvent),
		OperationType: b.operationType(rawEvent),
		Type:          b.eventType(rawEvent),
		OriginalRaw:   data,
	}, nil
}

func (*InventoryEventBuilder) primaryKeys(event InventoryEvent) entities.PkFields {
	return entities.PkFields{
		{Key: "resourceName", Value: event.Asset.Name},
		{Key: "resourceType", Value: event.Asset.AssetType},
	}
}

func (*InventoryEventBuilder) operationType(event InventoryEvent) entities.Operation {
	switch {
	case event.Deleted:
		return entities.Delete
	case event.PriorAssetState == "DOES_NOT_EXIST":
		return entities.Write
	case event.PriorAssetState == "PRESENT":
		return entities.Write
	default:
		return entities.Write
	}
}

func (*InventoryEventBuilder) eventType(event InventoryEvent) string {
	return event.Asset.AssetType
}

type InventoryEventAsset struct {
	Ancestors  []string               `json:"ancestors"`
	AssetType  string                 `json:"assetType"`
	Name       string                 `json:"name"`
	Resource   map[string]interface{} `json:"resource"`
	UpdateTime string                 `json:"updateTime"`
}

type InventoryEventWindow struct {
	StartTime string `json:"startTime"`
}

type InventoryEvent struct {
	Asset           InventoryEventAsset  `json:"asset"`
	PriorAsset      InventoryEventAsset  `json:"priorAsset"`
	PriorAssetState string               `json:"priorAssetState"`
	Window          InventoryEventWindow `json:"window"`
	Deleted         bool                 `json:"deleted"`
}
