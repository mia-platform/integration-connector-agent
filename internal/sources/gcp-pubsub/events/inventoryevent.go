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
	"github.com/mia-platform/integration-connector-agent/entities"
)

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

func (e InventoryEvent) Name() string {
	return e.Asset.Name
}

func (e InventoryEvent) AssetType() string {
	return e.Asset.AssetType
}

func (e InventoryEvent) Operation() entities.Operation {
	switch {
	case e.Deleted:
		return entities.Delete
	case e.PriorAssetState == "DOES_NOT_EXIST":
		return entities.Write
	case e.PriorAssetState == "PRESENT":
		return entities.Write
	default:
		return entities.Write
	}
}

func (e InventoryEvent) EventType() string {
	return RealtimeSyncEventType
}

func (e InventoryEvent) Ancestors() []string {
	return e.Asset.Ancestors
}

type InventoryImportEvent struct {
	AssetName string
	Type      string
}

func (e InventoryImportEvent) Name() string {
	return e.AssetName
}
func (e InventoryImportEvent) AssetType() string {
	return e.Type
}
func (e InventoryImportEvent) Operation() entities.Operation {
	return entities.Write
}
func (e InventoryImportEvent) EventType() string {
	return ImportEventType
}
func (e InventoryImportEvent) Ancestors() []string {
	// TODO: find a way to get ancestors for import events on each resource type
	return []string{}
}
