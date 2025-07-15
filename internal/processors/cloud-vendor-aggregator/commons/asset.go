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
