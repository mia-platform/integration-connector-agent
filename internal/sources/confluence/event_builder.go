// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package confluence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/entities"
)

//nolint:tagliatelle // Confluence import event uses snake_case, must maintain compatibility
type ConfluenceImportEvent struct {
	Type     string      `json:"type"`
	ID       string      `json:"id"`
	Key      string      `json:"key"`
	Name     string      `json:"name"`
	BaseURL  string      `json:"base_url"`
	SpaceKey string      `json:"space_key,omitempty"`
	Data     interface{} `json:"data"`
}

type ConfluenceEventBuilder struct{}

func NewConfluenceEventBuilder() *ConfluenceEventBuilder {
	return &ConfluenceEventBuilder{}
}

func (b *ConfluenceEventBuilder) GetPipelineEvent(ctx context.Context, data []byte) (entities.PipelineEvent, error) {
	var importEvent ConfluenceImportEvent
	if err := json.Unmarshal(data, &importEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Confluence import event: %w", err)
	}

	// Create a normalized payload structure that matches what the mapper expects
	normalizedEvent := map[string]interface{}{
		"type":     importEvent.Type,
		"id":       importEvent.ID,
		"key":      importEvent.Key,
		"name":     importEvent.Name,
		"base_url": importEvent.BaseURL,
		"data":     importEvent.Data,
	}

	// Add space key if present
	if importEvent.SpaceKey != "" {
		normalizedEvent["space_key"] = importEvent.SpaceKey
	}

	// Marshal the normalized event as the new payload
	normalizedData, err := json.Marshal(normalizedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal normalized import event: %w", err)
	}

	// Create the pipeline event with normalized import event data
	event := &entities.Event{
		PrimaryKeys: entities.PkFields{
			{Key: "id", Value: importEvent.ID},
		},
		Type:          importEvent.Type,
		OperationType: entities.Write, // All import events are treated as write operations
		OriginalRaw:   normalizedData,
	}

	return event, nil
}
