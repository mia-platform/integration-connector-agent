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

package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	t.Run("open server on port 3000", func(t *testing.T) {
		shutdown := make(chan interface{}, 1)

		envVars := config.EnvironmentVariables{
			HTTPPort:             "3000",
			HTTPAddress:          "127.0.0.1",
			LogLevel:             "error",
			DelayShutdownSeconds: 10,
		}
		cfg := &config.Configuration{}

		ctx := t.Context()
		go func() {
			assert.NoError(t, New(ctx, envVars, cfg, shutdown))
			assert.ErrorIs(t, ctx.Err(), context.Canceled)
		}()

		defer func() {
			shutdown <- struct{}{}
			close(shutdown)
		}()

		time.Sleep(1 * time.Second)
		resp, err := http.DefaultClient.Get("http://localhost:3000/-/healthz")
		require.NoError(t, err)

		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("sets correct path prefix", func(t *testing.T) {
		shutdown := make(chan interface{}, 1)

		envVars := config.EnvironmentVariables{
			HTTPPort:             "8080",
			HTTPAddress:          "127.0.0.1",
			ServicePrefix:        "/prefix",
			LogLevel:             "error",
			DelayShutdownSeconds: 10,
		}
		cfg := &config.Configuration{}
		go func() {
			assert.NoError(t, New(t.Context(), envVars, cfg, shutdown))
		}()
		defer func() { shutdown <- struct{}{} }()

		time.Sleep(1 * time.Second)
		resp, err := http.DefaultClient.Get("http://localhost:8080/prefix/")
		require.NoError(t, err)

		resp.Body.Close()
		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestShutdown(t *testing.T) {
	cfg := &config.Configuration{}
	shutdown := make(chan interface{}, 1)
	done := make(chan bool, 1)

	go func() {
		time.Sleep(5 * time.Second)
		done <- false
	}()

	go func() {
		envVars := config.EnvironmentVariables{
			HTTPAddress:          "127.0.0.1",
			HTTPPort:             "8080",
			LogLevel:             "error",
			DelayShutdownSeconds: 3,
		}
		assert.NoError(t, New(t.Context(), envVars, cfg, shutdown))
		done <- true
	}()

	shutdown <- struct{}{}

	flag := <-done
	assert.True(t, flag)
}
