package middleware

import (
	"errors"

	"latihan-fiber/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		const prefix = "Bearer "
		header := c.Get("Authorization")
		if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "token tidak ditemukan")
		}
		user, err := jwtManager.Parse(header[len(prefix):])
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			if errors.Is(err, jwt.ErrTokenExpired) {
				return helper.Fail(c, fiber.StatusUnauthorized, "token kedaluwarsa")
			}
			return helper.Fail(c, fiber.StatusUnauthorized, "token tidak valid")
		}
		c.Locals(helper.LocalsAuthUser, user)
		return c.Next()
	}
}
