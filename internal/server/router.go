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
	"os"
	"path"
	"path/filepath"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	swagger "github.com/davidebianchi/gswagger"
	oasfiber "github.com/davidebianchi/gswagger/support/fiber"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
	middleware "github.com/mia-platform/glogger/v4/middleware/fiber"
	"github.com/sirupsen/logrus"
)

func NewApp(ctx context.Context, env config.EnvironmentVariables, log *logrus.Logger, cfg *config.Configuration) (*fiber.App, error) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	cmdName := filepath.Base(os.Args[0])
	middlewareLog := glogrus.GetLogger(logrus.NewEntry(log))
	app.Use(middleware.RequestMiddlewareLogger(middlewareLog, []string{"/-/"}))
	statusRoutes(app, cmdName, utils.ServiceVersionInformation())
	if env.ServicePrefix != "" && env.ServicePrefix != "/" {
		log.WithField("servicePrefix", env.ServicePrefix).Trace("applying service prefix")
		app.Use(pprof.New(pprof.Config{Prefix: path.Clean(env.ServicePrefix)}))
	}

	oasRouter, err := swagger.NewRouter(oasfiber.NewRouter(app), swagger.Options{
		Context: context.Background(),
		Openapi: &openapi3.T{
			Info: &openapi3.Info{
				Title:   cmdName,
				Version: utils.Version,
			},
		},
		JSONDocumentationPath: "/documentations/json",
		YAMLDocumentationPath: "/documentations/yaml",
		PathPrefix:            env.ServicePrefix,
	})
	if err != nil {
		return nil, err
	}

	closeHandler, err := setupPipelines(ctx, log, cfg, oasRouter)
	if err != nil {
		return nil, err
	}
	app.Hooks().OnShutdown(closeHandler)

	if err := oasRouter.GenerateAndExposeOpenapi(); err != nil {
		return nil, err
	}

	return app, nil
}
