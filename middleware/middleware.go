package middleware

import (
	"strings"

	"latihan-fiber/app/model"
	"latihan-fiber/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
	app.Use(logger.New(logger.Config{
		Output: config.LoggerWriter(),
		Format: "{\"time\":\"${time}\",\"request_id\":\"${locals:requestid}\",\"method\":\"${method}\",\"path\":\"${path}\",\"status\":${status},\"latency\":\"${latency}\"}\n",
	}))
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
