package api

import (
	"web-crawler/middleware"
	"web-crawler/service"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type Api struct {
	logger *logrus.Logger

	service *service.Service
}

func NewApi(
	logger *logrus.Logger,
	service *service.Service,
) *Api {
	return &Api{
		logger: logger,

		service: service,
	}
}

func (api *Api) SetupRoutes(app *fiber.App) *fiber.App {
	// Error handler middleware
	app.Use(middleware.ErrorHandler())

	// Emas Routes
	emas := app.Group("/emas")
	emas.Get("/", api.GetAllEmas)

	return app
}
