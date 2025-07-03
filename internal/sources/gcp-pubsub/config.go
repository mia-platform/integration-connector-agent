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

package gcppubsub

import (
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/config"
)

type Config struct {
	ProjectID          string              `json:"projectId"`
	TopicName          string              `json:"topicName"`
	SubscriptionID     string              `json:"subscriptionId"`
	AckDeadlineSeconds int                 `json:"ackDeadlineSeconds,omitempty"`
	CredentialsJSON    config.SecretSource `json:"credentialsJson,omitempty"`
}

func (c *Config) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("projectId must be provided")
	}
	if c.TopicName == "" {
		return fmt.Errorf("topicName must be provided")
	}
	if c.SubscriptionID == "" {
		return fmt.Errorf("subscriptionId must be provided")
	}

	return nil
}
