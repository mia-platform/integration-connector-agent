// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
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

package jboss

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			WildFlyURL:      "http://localhost:9990/management",
			Username:        "admin",
			Password:        "password",
			PollingInterval: Duration(30 * time.Second),
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing password", func(t *testing.T) {
		cfg := &Config{
			WildFlyURL: "http://localhost:9990/management",
			Username:   "admin",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password must be provided")
	})
}

func TestConfig_withDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.withDefaults()

	assert.Equal(t, defaultWildFlyURL, cfg.WildFlyURL)
	assert.Equal(t, defaultUsername, cfg.Username)
	assert.Equal(t, Duration(defaultPollingInterval), cfg.PollingInterval)
}

func TestCreateDeploymentEvent(t *testing.T) {
	deployment := Deployment{
		Name:        "test-app.war",
		RuntimeName: "test-app.war",
		Status:      "OK",
		Enabled:     true,
	}

	timestamp := time.Now()

	event, err := CreateDeploymentEvent(deployment, timestamp)
	require.NoError(t, err)

	primaryKeys := event.GetPrimaryKeys()
	assert.Len(t, primaryKeys, 1)
	assert.Equal(t, "deploymentName", primaryKeys[0].Key)
	assert.Equal(t, "test-app.war", primaryKeys[0].Value)
	assert.NotNil(t, event.Data())
	assert.Equal(t, DeploymentStatusEvent, event.GetType())
}

func TestJBossEventBuilder_GetPipelineEvent(t *testing.T) {
	builder := NewJBossEventBuilder()

	deploymentEvent := DeploymentEvent{
		Deployment: Deployment{
			Name:    "test-app.war",
			Status:  "OK",
			Enabled: true,
		},
		Timestamp: time.Now(),
		EventType: DeploymentStatusEvent,
	}

	data, err := json.Marshal(deploymentEvent)
	require.NoError(t, err)

	event, err := builder.GetPipelineEvent(t.Context(), data)
	require.NoError(t, err)

	primaryKeys := event.GetPrimaryKeys()
	assert.Len(t, primaryKeys, 1)
	assert.Equal(t, "deploymentName", primaryKeys[0].Key)
	assert.Equal(t, "test-app.war", primaryKeys[0].Value)
}
