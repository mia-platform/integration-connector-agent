// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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
	// We always write the event regardless of status
	operation := entities.Write

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
