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

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v3"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
)

type Client interface {
	GetByID(resourceID, apiVersion string) (*Resource, error)
}

func NewClient(credentials azcore.TokenCredential) Client {
	return &azureClient{
		credentials: credentials,
	}
}

type azureClient struct {
	credentials azcore.TokenCredential
}

type Resource struct {
	Name     string
	Tags     commons.Tags
	Type     string
	Location string
}

func (a *azureClient) GetByID(resourceID, apiVersion string) (*Resource, error) {
	genericClient, err := armresources.NewClient("", a.credentials, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure resources client: %w", err)
	}

	res, err := genericClient.GetByID(context.Background(), resourceID, apiVersion, nil)
	if err != nil {
		return nil, err
	}

	tags := make(commons.Tags)
	for key, value := range res.Tags {
		tags[key] = str(value)
	}

	return &Resource{
		Name:     str(res.Name),
		Type:     str(res.Type),
		Location: str(res.Location),
		Tags:     tags,
	}, nil
}

func str(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
