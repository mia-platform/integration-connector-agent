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

package utils

import (
	"encoding/base64"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// MinBase64Length is the minimum length for a valid base64 string
	MinBase64Length = 4
	// MaxBodySizeForDecoding is the maximum body size we'll attempt to decode
	MaxBodySizeForDecoding = 10000
)

func IsNil(i any) bool {
	defer func() { recover() }() //nolint:errcheck
	return i == nil || reflect.ValueOf(i).IsNil()
}

// IsBase64 checks if a string appears to be base64 encoded
func IsBase64(s string) bool {
	// Must be at least 4 characters long for valid base64
	if len(s) < MinBase64Length {
		return false
	}

	// Should be multiple of 4 characters (with padding)
	if len(s)%4 != 0 {
		return false
	}

	// Check for valid base64 characters
	matched, _ := regexp.MatchString(`^[A-Za-z0-9+/]*={0,2}$`, s)
	if !matched {
		return false
	}

	// Try to decode it
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// DecodeBase64IfValid attempts to decode a string if it appears to be base64
// Returns the decoded string and whether it was successfully decoded
func DecodeBase64IfValid(s string) (string, bool) {
	if !IsBase64(s) {
		return s, false
	}

	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s, false
	}

	// Check if decoded bytes are valid UTF-8
	if !utf8.Valid(decoded) {
		return s, false
	}

	return string(decoded), true
}

// TryDecodeBase64Body attempts to decode the body content if it's base64-encoded
// Returns the original and decoded content (if applicable) for logging
func TryDecodeBase64Body(body []byte) (original string, decoded string, wasDecoded bool) {
	original = string(body)

	// Skip very large bodies to avoid performance issues
	if len(body) > MaxBodySizeForDecoding {
		return original, "", false
	}

	// Try to decode as base64
	bodyStr := strings.TrimSpace(string(body))
	if decodedStr, success := DecodeBase64IfValid(bodyStr); success {
		return original, decodedStr, true
	}

	return original, "", false
}
