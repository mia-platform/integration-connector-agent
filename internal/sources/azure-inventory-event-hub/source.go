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

package azureinventoryeventhub

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	azureeventhub "github.com/mia-platform/integration-connector-agent/internal/sources/azure-event-hub"
	"github.com/sirupsen/logrus"
)

func AddSource(ctx context.Context, cfg config.GenericConfig, pg pipeline.IPipelineGroup, logger *logrus.Logger) error {
	eventHubConfig, err := config.GetConfig[*azureeventhub.Config](cfg)
	if err != nil {
		return err
	}

	eventHubConfig.EventConsumer = func(_ *azeventhubs.ReceivedEventData) error {
		return nil
	}

	if err := eventHubConfig.Validate(); err != nil {
		return fmt.Errorf("invalid event hub configuration: %w", err)
	}

	return azureeventhub.SetupEventHub(ctx, eventHubConfig, pg, logger)
}
