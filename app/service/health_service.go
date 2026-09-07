package service

import (
	"context"
	"time"

	"latihan-fiber/app/model"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Health(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		pingCtx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(pingCtx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(model.WebResponse{
				Success: false,
				Message: "database tidak tersedia",
			})
		}
		return c.JSON(model.WebResponse{Success: true, Message: "service sehat"})
	}
}
