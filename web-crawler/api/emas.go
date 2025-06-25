package api

import (
	"fmt"

	"web-crawler/service"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func (api *Api) GetAllEmas(c *fiber.Ctx) error {
	const op = "[api] - Api.GetAllEmas"

	// Parse request queries
	page := c.QueryInt("page", 1)
	size := c.QueryInt("size", 10)

	params := &service.GetAllEmasParams{
		Page: int32(page),
		Size: int32(size),
	}

	logger := api.logger.WithFields(logrus.Fields{
		"[op]":   op,
		"params": fmt.Sprintf("%+v", params),
	})

	logger.Info()

	result, err := api.service.GetAllEmas(c.Context(), params)
	if err != nil {
		logger.WithError(err).Error()

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
