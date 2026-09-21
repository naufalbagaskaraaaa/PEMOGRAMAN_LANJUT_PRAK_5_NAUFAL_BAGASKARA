package helper

import (
	"latihan-fiber/app/model"

	"github.com/gofiber/fiber/v2"
)

func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	if c == nil {
		return model.AuthUser{}, false
	}
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}
