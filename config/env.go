// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package config

// EnvironmentVariables struct with the mapping of desired
// environment variables.
type EnvironmentVariables struct {
	LogLevel             string `env:"LOG_LEVEL" envDefault:"info"`
	HTTPPort             string `env:"HTTP_PORT" envDefault:"8080"`
	ServicePrefix        string `env:"SERVICE_PREFIX"`
	ServiceVersion       string `env:"SERVICE_VERSION"`
	DelayShutdownSeconds int    `env:"DELAY_SHUTDOWN_SECONDS" envDefault:"10"`
}
