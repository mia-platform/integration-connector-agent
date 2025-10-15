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

package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitHubClient(t *testing.T) {
	client, err := NewGitHubClient("test-token", "test-org")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "test-token", client.token)
	assert.Equal(t, "test-org", client.organization)
	assert.Equal(t, "token", client.authType)
	assert.Equal(t, "https://api.github.com", client.baseURL)
}

func TestNewGitHubClientWithApp(t *testing.T) {
	// This test will fail in CI because it tries to get a real OAuth token
	// but it demonstrates the structure
	t.Skip("Skipping OAuth test - requires real GitHub App credentials")

	client, err := NewGitHubClientWithApp("test-client-id", "test-client-secret", "test-org")
	if err != nil {
		// Expected to fail with invalid credentials
		t.Logf("Expected OAuth error: %v", err)
		return
	}

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "test-client-id", client.clientID)
	assert.Equal(t, "test-client-secret", client.clientSecret)
	assert.Equal(t, "test-org", client.organization)
	assert.Equal(t, "app", client.authType)
	assert.Equal(t, "https://api.github.com", client.baseURL)
	assert.NotEmpty(t, client.accessToken)
}

func TestGitHubClientAuthType(t *testing.T) {
	// Test token-based client
	tokenClient, err := NewGitHubClient("test-token", "test-org")
	require.NoError(t, err)
	assert.Equal(t, "token", tokenClient.authType)

	// We can't test the app client without real credentials,
	// but we can verify the structure is correct
	appClient := &GitHubClient{
		clientID:     "test-id",
		clientSecret: "test-secret",
		accessToken:  "test-access-token",
		organization: "test-org",
		baseURL:      "https://api.github.com",
		authType:     "app",
	}
	assert.Equal(t, "app", appClient.authType)
	assert.Equal(t, "test-access-token", appClient.accessToken)
}
