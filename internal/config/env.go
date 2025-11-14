// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package config

// EnvironmentVariables struct with the mapping of desired
// environment variables.
type EnvironmentVariables struct {
	LogLevel             string `env:"LOG_LEVEL" envDefault:"info"`
	HTTPPort             string `env:"HTTP_PORT" envDefault:"8080"`
	HTTPAddress          string `env:"HTTP_ADDRESS" envDefault:"0.0.0.0"`
	ServicePrefix        string `env:"SERVICE_PREFIX"`
	DelayShutdownSeconds int    `env:"DELAY_SHUTDOWN_SECONDS" envDefault:"10"`

	ConfigurationPath string `env:"CONFIGURATION_PATH,required"`
}
