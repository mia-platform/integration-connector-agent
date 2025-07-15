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

package cloudvendoraggregator

import (
	"github.com/mia-platform/integration-connector-agent/entities"
	awsaggergator "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws"
	azureaggregator "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/azure"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"
	gcpaggregator "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp"

	"github.com/sirupsen/logrus"
)

type CloudVendorAggregator struct {
}

func New(logger *logrus.Logger, cfg config.Config) (entities.Processor, error) {
	switch cfg.CloudVendorName {
	case commons.GCPAssetProvider:
		return gcpaggregator.New(logger, cfg.AuthOptions)
	case commons.AWSAssetProvider:
		return awsaggergator.New(logger, cfg.AuthOptions), nil
	case commons.AzureAssetProvider:
		return azureaggregator.New(logger, cfg.AuthOptions)
	default:
		return nil, config.ErrInvalidCloudVendor
	}
}
