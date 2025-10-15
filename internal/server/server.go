// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/mia-platform/integration-connector-agent/internal/config"

	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
)

func New[Signal any](ctx context.Context, envVars config.EnvironmentVariables, cfg *config.Configuration, sysChannel <-chan Signal) error {
	// Init logger instance.
	ctxWithCancel, cancel := context.WithCancel(ctx)
	log, err := glogrus.InitHelper(glogrus.InitOptions{Level: envVars.LogLevel})
	if err != nil {
		panic(err)
	}

	app, err := NewApp(ctxWithCancel, envVars, log, cfg)
	if err != nil {
		cancel()
		return err
	}

	go func() {
		log.WithField("port", envVars.HTTPPort).Info("starting server")
		if err := app.Listen(fmt.Sprintf("%s:%s", envVars.HTTPAddress, envVars.HTTPPort)); err != nil {
			log.Println(err)
		}
	}()

	<-sysChannel
	time.Sleep(time.Duration(envVars.DelayShutdownSeconds) * time.Second)
	log.Info("Gracefully shutting down...")

	cancel() // shutting down server, cancel the context
	if err := app.Shutdown(); err != nil {
		return err
	}

	return nil
}
