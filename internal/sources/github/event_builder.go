// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mia-platform/integration-connector-agent/entities"
)

//nolint:tagliatelle // GitHub API uses snake_case, must maintain compatibility
type GitHubImportEvent struct {
	Type         string      `json:"type"`
	ID           int64       `json:"id"`
	Name         string      `json:"name"`
	FullName     string      `json:"full_name"`
	Organization string      `json:"organization"`
	Repository   string      `json:"repository,omitempty"`
	Data         interface{} `json:"data"`
}

type GitHubEventBuilder struct{}

func NewGitHubEventBuilder() *GitHubEventBuilder {
	return &GitHubEventBuilder{}
}

func (b *GitHubEventBuilder) GetPipelineEvent(ctx context.Context, data []byte) (entities.PipelineEvent, error) {
	var importEvent GitHubImportEvent
	if err := json.Unmarshal(data, &importEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GitHub import event: %w", err)
	}

	// Create a normalized payload structure that matches what the mapper expects
	normalizedEvent := map[string]interface{}{
		"type":         importEvent.Type,
		"id":           importEvent.ID,
		"name":         importEvent.Name,
		"full_name":    importEvent.FullName,
		"organization": importEvent.Organization,
		"repository":   importEvent.Repository,
		"data":         importEvent.Data,
	}

	// Marshal the normalized event as the new payload
	normalizedData, err := json.Marshal(normalizedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal normalized import event: %w", err)
	}

	// Create the pipeline event with normalized import event data
	event := &entities.Event{
		PrimaryKeys: entities.PkFields{
			{Key: "id", Value: strconv.FormatInt(importEvent.ID, 10)},
		},
		Type:          importEvent.Type,
		OperationType: entities.Write, // All import events are treated as write operations
		OriginalRaw:   normalizedData,
	}

	return event, nil
}
