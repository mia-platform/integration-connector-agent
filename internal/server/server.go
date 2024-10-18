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
	"fmt"
	"os"
	"time"

	"github.com/mia-platform/data-connector-agent/internal/config"
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
)

func New(envVars config.EnvironmentVariables, sysChannel <-chan os.Signal) error {
	// Init logger instance.
	log, err := glogrus.InitHelper(glogrus.InitOptions{Level: envVars.LogLevel})
	if err != nil {
		panic(err)
	}

	app, err := NewRouter(envVars, log)
	if err != nil {
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
	if err := app.Shutdown(); err != nil {
		return err
	}

	return nil
}