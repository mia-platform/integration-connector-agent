// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package entities

import (
	"context"
	"encoding/gob"
	"encoding/json"
)

func init() {
	gob.RegisterName("entities.Event", Event{})
}

//go:generate ${TOOLS_BIN}/stringer -type=Operation
type Operation int

const (
	Write Operation = iota
	Delete
)

type PkField struct {
	Key   string
	Value string
}

type PkFields []PkField

func (fields PkFields) Map() map[string]string {
	m := map[string]string{}
	for _, f := range fields {
		m[f.Key] = f.Value
	}
	return m
}

func (fields PkFields) IsEmpty() bool {
	return len(fields) == 0
}

type PipelineEvent interface {
	GetPrimaryKeys() PkFields
	GetType() string

	Data() []byte
	Operation() Operation
	WithData([]byte)
	JSON() (map[string]any, error)
	Clone() PipelineEvent
}

type EventBuilder interface {
	GetPipelineEvent(ctx context.Context, data []byte) (PipelineEvent, error)
}

type Event struct {
	PrimaryKeys   PkFields
	Type          string
	OperationType Operation

	OriginalRaw []byte
	jsonData    map[string]any
}

func (e Event) GetPrimaryKeys() PkFields {
	return e.PrimaryKeys
}

func (e Event) Data() []byte {
	return e.OriginalRaw
}

func (e Event) Operation() Operation {
	return e.OperationType
}

func (e Event) JSON() (map[string]any, error) {
	if e.jsonData != nil {
		return e.jsonData, nil
	}
	parsed := map[string]any{}

	if err := json.Unmarshal(e.OriginalRaw, &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

func (e *Event) WithData(raw []byte) {
	e.jsonData = nil
	e.OriginalRaw = raw
}

func (e *Event) Clone() PipelineEvent {
	return &Event{
		PrimaryKeys:   e.PrimaryKeys,
		OperationType: e.OperationType,
		OriginalRaw:   e.OriginalRaw,
		Type:          e.Type,

		jsonData: e.jsonData,
	}
}

func (e Event) GetType() string {
	return e.Type
}
