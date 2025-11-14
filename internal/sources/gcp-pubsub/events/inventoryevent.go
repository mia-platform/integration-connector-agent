// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package gcppubsubevents

import (
	"cloud.google.com/go/asset/apiv1/assetpb"
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
	Data      *assetpb.Asset
}

func (e InventoryImportEvent) Name() string {
	return e.AssetName
}
func (e InventoryImportEvent) AssetType() string {
	return e.Type
}
func (e InventoryImportEvent) AssetData() *assetpb.Asset {
	return e.Data
}
func (e InventoryImportEvent) Operation() entities.Operation {
	return entities.Write
}
func (e InventoryImportEvent) EventType() string {
	return e.Type
}
func (e InventoryImportEvent) Ancestors() []string {
	// TODO: find a way to get ancestors for import events on each resource type
	return []string{}
}
