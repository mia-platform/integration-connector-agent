// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadSecret(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		source         secretConfig
		expectedSecret string
	}{
		"from missing env return empty string": {
			source: secretConfig{
				FromEnv: "ENV_NAME",
			},
		},
		"from missing file return empty string": {
			source: secretConfig{
				FromFile: filepath.Join("testdata", "secretsource", "missing"),
			},
		},
		"from missing secret section return emptry string": {
			source: secretConfig{},
		},
		"from valid file return secret string": {
			source: secretConfig{
				FromFile: filepath.Join("testdata", "secretsource", "secret"),
			},
			expectedSecret: "secret-from-file",
		},
		"with both from env has priority": {
			source: secretConfig{
				FromEnv:  "ENV_NAME",
				FromFile: filepath.Join("testdata", "secretsource", "secret"),
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			secret := readSecret(&test.source)
			require.Equal(t, test.expectedSecret, secret)
		})
	}
}

func TestSecretSource(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		source         string
		expectedSecret string
		expectedError  string
	}{
		"from missing env return empty string": {
			source: `{"fromEnv": "ENV_NAME"}`,
		},
		"from missing file return empty string": {
			source: fmt.Sprintf(`{"fromFile": "%s"}`, filepath.Join("testdata", "secretsource", "missing")),
		},
		"from missing secret section return empty string": {
			source: `{}`,
		},
		"from valid file return secret string": {
			source:         fmt.Sprintf(`{"fromFile": "%s"}`, filepath.Join("testdata", "secretsource", "secret")),
			expectedSecret: "secret-from-file",
		},
		"with both from env has priority": {
			source: fmt.Sprintf(`{"fromEnv": "ENV_NAME", "fromFile": "%s"}`, filepath.Join("testdata", "secretsource", "secret")),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			var actual SecretSource
			err := json.Unmarshal([]byte(test.source), &actual)
			require.NoError(t, err)
			require.Equal(t, test.expectedSecret, actual.String())
		})
	}
}
