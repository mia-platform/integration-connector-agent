// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

	integrations, err := setupIntegrations(ctx, log, cfg, oasRouter)
	if err != nil {
		return nil, err
	}

	if err := oasRouter.GenerateAndExposeOpenapi(); err != nil {
		return nil, err
	}

	go func(integrations []*Integration) {
		ch := ctx.Done()
		<-ch

		for _, integration := range integrations {
			integration.Close(ctx)
		}
	}(integrations)
	return app, nil
}
