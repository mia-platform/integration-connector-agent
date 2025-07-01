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

package awssqs

import (
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/config"
)

type Config struct {
	QueueURL        string              `json:"queueUrl"`
	Region          string              `json:"region"`
	AccessKeyID     string              `json:"accessKeyId,omitempty"`
	SecretAccessKey config.SecretSource `json:"secretAccessKey,omitempty"`
	SessionToken    config.SecretSource `json:"sessionToken,omitempty"`
}

func (c *Config) Validate() error {
	if c.QueueURL == "" {
		return fmt.Errorf("queueId must be provided")
	}

	if c.Region == "" {
		return fmt.Errorf("region must be provided")
	}

	if c.AccessKeyID == "" {
		return fmt.Errorf("accessKeyId must be provided")
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("secretAccessKey must be provided")
	}

	return nil
}
