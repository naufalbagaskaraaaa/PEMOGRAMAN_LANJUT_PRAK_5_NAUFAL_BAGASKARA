package middleware

import (
	"latihan-fiber/helper"

	"github.com/gofiber/fiber/v2"
)

func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
		}
		if !perms.Can(user.Role, permission) {
			return helper.Fail(c, fiber.StatusForbidden, "role "+user.Role+" tidak memiliki hak "+permission)
		}
		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	allowedRoles := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowedRoles[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
		}
		if _, allowed := allowedRoles[user.Role]; !allowed {
			return helper.Fail(c, fiber.StatusForbidden, "role "+user.Role+" tidak diizinkan")
		}
		return c.Next()
	}
}
