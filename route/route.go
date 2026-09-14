package route

import (
	"latihan-fiber/app/service"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(app *fiber.App, pool *pgxpool.Pool, studentService *service.StudentService, authService *service.AuthService, jwtManager *helper.JWTManager) {
	app.Get("/health", service.Health(pool))

	api := app.Group("/api/v1")
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", authService.Register)
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)
	auth.Get("/me", middleware.RequireAuth(jwtManager), authService.Me)

	students := api.Group("/students", middleware.RequireAuth(jwtManager), middleware.RequireJSON)
	students.Get("/", studentService.List)
	students.Get("/:id", studentService.Get)
	students.Post("/", studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", studentService.Delete)
}
