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

package jboss

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mia-platform/integration-connector-agent/entities"
)

const (
	// Event types for JBoss/WildFly deployments
	DeploymentStatusEvent = "jboss:deployment_status"

	// Primary key path for deployment events
	deploymentEventIDPath = "deployment.name"
)

// DeploymentEvent represents a JBoss/WildFly deployment status event
type DeploymentEvent struct {
	Deployment Deployment `json:"deployment"`
	Timestamp  time.Time  `json:"timestamp"`
	EventType  string     `json:"eventType"`
}

// CreateDeploymentEvent creates a pipeline event from a deployment
func CreateDeploymentEvent(deployment Deployment, timestamp time.Time) (entities.PipelineEvent, error) {
	deploymentEvent := DeploymentEvent{
		Deployment: deployment,
		Timestamp:  timestamp,
		EventType:  DeploymentStatusEvent,
	}

	data, err := json.Marshal(deploymentEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deployment event: %w", err)
	}

	// Determine operation based on deployment status
	var operation entities.Operation
	if deployment.Enabled && deployment.Status == "OK" {
		operation = entities.Write
	} else {
		operation = entities.Write // We still write the event but with different status
	}

	// Create primary key fields from deployment name
	primaryKeys := entities.PkFields{
		{Key: "deploymentName", Value: deployment.Name},
	}

	event := &entities.Event{
		PrimaryKeys:   primaryKeys,
		Type:          DeploymentStatusEvent,
		OperationType: operation,
		OriginalRaw:   data,
	}

	return event, nil
}

// JBossEventBuilder implements the EventBuilder interface for JBoss events
type JBossEventBuilder struct{}

// NewJBossEventBuilder creates a new JBoss event builder
func NewJBossEventBuilder() *JBossEventBuilder {
	return &JBossEventBuilder{}
}

// GetPipelineEvent creates a pipeline event from raw data
func (b *JBossEventBuilder) GetPipelineEvent(ctx context.Context, data []byte) (entities.PipelineEvent, error) {
	var deploymentEvent DeploymentEvent
	if err := json.Unmarshal(data, &deploymentEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JBoss deployment event: %w", err)
	}

	return CreateDeploymentEvent(deploymentEvent.Deployment, deploymentEvent.Timestamp)
}
