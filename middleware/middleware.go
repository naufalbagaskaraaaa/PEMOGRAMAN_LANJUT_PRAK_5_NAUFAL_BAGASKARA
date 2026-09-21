package middleware

import (
	"log/slog"
	"strings"
	"time"

	"latihan-fiber/app/model"
	"latihan-fiber/config"
	"latihan-fiber/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

var bodyMethods = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func RequireJSON(c *fiber.Ctx) error {
	if bodyMethods[c.Method()] && !strings.HasPrefix(c.Get("Content-Type"), fiber.MIMEApplicationJSON) {
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(model.WebResponse{
			Success: false,
			Message: "Content-Type harus application/json",
		})
	}
	return c.Next()
}

func Register(app *fiber.App) {
	app.Use(requestid.New())
	logger := slog.New(slog.NewJSONHandler(config.LoggerWriter(), nil))
	app.Use(func(c *fiber.Ctx) error {
		started := time.Now()
		err := c.Next()
		requestID, _ := c.Locals("requestid").(string)

		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("latency", time.Since(started)),
		}
		if user, ok := helper.CurrentUser(c); ok {
			attrs = append(attrs, slog.Int("user_id", user.UserID), slog.String("role", user.Role))
		}
		logger.Info("http_request", attrs...)
		return err
	})
	app.Use(cors.New())
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	message := "terjadi kesalahan internal server"
	if fiberErr, ok := err.(*fiber.Error); ok {
		status = fiberErr.Code
		message = fiberErr.Message
	}
	return c.Status(status).JSON(model.WebResponse{Success: false, Message: message})
}

func NotFound(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(model.WebResponse{
		Success: false,
		Message: "endpoint tidak ditemukan",
	})
}
