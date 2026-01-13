// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package azure

import (
	"encoding/json"

	"github.com/mia-platform/integration-connector-agent/entities"
)

//go:generate ${TOOLS_BIN}/stringer -type=EventType -trimprefix=EventType
type EventType int

const (
	EventTypeRecordFromEventHub EventType = iota
)

func EventFromRecord(record *ActivityLogEventRecord) *entities.Event {
	rawRecord, err := json.Marshal(record)
	if err != nil {
		return nil
	}

	return &entities.Event{
		PrimaryKeys:   primaryKeys(record.ResourceID),
		OperationType: record.entityOperationType(),
		Type:          EventTypeRecordFromEventHub.String(),
		OriginalRaw:   rawRecord,
	}
}
