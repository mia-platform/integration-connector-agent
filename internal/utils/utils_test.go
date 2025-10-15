// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package utils

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsNil(t *testing.T) {
	var i interface{}
	require.True(t, IsNil(i))

	type fooType struct{}
	var foo *fooType
	require.True(t, IsNil(foo))

	require.False(t, IsNil(fooType{}))

	var s []string
	require.True(t, IsNil(s))
}

func TestIsBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid base64 string",
			input:    base64.StdEncoding.EncodeToString([]byte("hello world")),
			expected: true,
		},
		{
			name:     "invalid base64 - too short",
			input:    "ab",
			expected: false,
		},
		{
			name:     "invalid base64 - wrong length",
			input:    "abcde",
			expected: false,
		},
		{
			name:     "invalid base64 - invalid characters",
			input:    "abc@",
			expected: false,
		},
		{
			name:     "plain text",
			input:    "hello world",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBase64(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecodeBase64IfValid(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedOutput  string
		expectedDecoded bool
	}{
		{
			name:            "valid base64 string",
			input:           base64.StdEncoding.EncodeToString([]byte("hello world")),
			expectedOutput:  "hello world",
			expectedDecoded: true,
		},
		{
			name:            "plain text",
			input:           "hello world",
			expectedOutput:  "hello world",
			expectedDecoded: false,
		},
		{
			name:            "invalid base64",
			input:           "abc@",
			expectedOutput:  "abc@",
			expectedDecoded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, decoded := DecodeBase64IfValid(tt.input)
			assert.Equal(t, tt.expectedOutput, output)
			assert.Equal(t, tt.expectedDecoded, decoded)
		})
	}
}

func TestTryDecodeBase64Body(t *testing.T) {
	tests := []struct {
		name               string
		input              []byte
		expectedOriginal   string
		expectedDecoded    string
		expectedWasDecoded bool
	}{
		{
			name:               "valid base64 body",
			input:              []byte(base64.StdEncoding.EncodeToString([]byte("hello world"))),
			expectedOriginal:   base64.StdEncoding.EncodeToString([]byte("hello world")),
			expectedDecoded:    "hello world",
			expectedWasDecoded: true,
		},
		{
			name:               "plain text body",
			input:              []byte("hello world"),
			expectedOriginal:   "hello world",
			expectedDecoded:    "",
			expectedWasDecoded: false,
		},
		{
			name:               "json body",
			input:              []byte(`{"key": "value"}`),
			expectedOriginal:   `{"key": "value"}`,
			expectedDecoded:    "",
			expectedWasDecoded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original, decoded, wasDecoded := TryDecodeBase64Body(tt.input)
			assert.Equal(t, tt.expectedOriginal, original)
			assert.Equal(t, tt.expectedDecoded, decoded)
			assert.Equal(t, tt.expectedWasDecoded, wasDecoded)
		})
	}
}
