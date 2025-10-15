// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package commons

import "time"

const (
	AWSAssetProvider   = "aws"
	AzureAssetProvider = "azure"
	GCPAssetProvider   = "gcp"
)

type Tags map[string]string

type Asset struct {
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Provider      string    `json:"provider"`
	Location      string    `json:"location"`
	Relationships []string  `json:"relationships"`
	Tags          Tags      `json:"tags"`
	RawData       []byte    `json:"rawData"`
	Timestamp     time.Time `json:"timestamp"`
}

func NewAsset(name, assetType, provider string) *Asset {
	return &Asset{
		Name:      name,
		Type:      assetType,
		Provider:  provider,
		Timestamp: time.Now(),
	}
}

func (a *Asset) WithLocation(location string) *Asset {
	a.Location = location
	return a
}
func (a *Asset) WithRelationships(relationships []string) *Asset {
	a.Relationships = relationships
	return a
}
func (a *Asset) WithTags(tags Tags) *Asset {
	a.Tags = tags
	return a
}
func (a *Asset) WithRawData(rawData []byte) *Asset {
	a.RawData = rawData
	return a
}
