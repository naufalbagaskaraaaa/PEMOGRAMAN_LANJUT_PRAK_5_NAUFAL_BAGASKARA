package routes

import (
	"latihan-fiber/app/service"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService interface {
	List(*fiber.Ctx) error
	Create(*fiber.Ctx) error
	Delete(*fiber.Ctx) error
	AssignRole(*fiber.Ctx) error
	Get(*fiber.Ctx) error
	Replace(*fiber.Ctx) error
	Patch(*fiber.Ctx) error
}

type Dependencies struct {
	Pool            *pgxpool.Pool
	StudentService  *service.StudentService
	PrestasiService *service.PrestasiService
	AuthService     *service.AuthService
	UserService     UserService
	JWT             *helper.JWTManager
	Permissions     *helper.PermissionSet
}

func Register(app *fiber.App, deps Dependencies) {
	app.Get("/health", service.Health(deps.Pool))

	api := app.Group("/api/v1")
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	students := api.Group("/students", middleware.RequireAuth(deps.JWT), middleware.RequireJSON)
	perms := deps.Permissions
	students.Get("/", middleware.RequirePermission(perms, "student:list"), deps.StudentService.List)
	students.Get("/:id", deps.StudentService.Get)
	students.Post("/", middleware.RequirePermission(perms, "student:create"), deps.StudentService.Create)
	students.Put("/:id", deps.StudentService.Replace)
	students.Patch("/:id", deps.StudentService.Patch)
	students.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), deps.StudentService.Delete)

	prestasi := api.Group("/prestasi", middleware.RequireAuth(deps.JWT), middleware.RequireJSON)
	prestasi.Post("/", deps.PrestasiService.Create)
	prestasi.Get("/:id_prestasi", deps.PrestasiService.Get)

	if deps.UserService == nil {
		return
	}

	users := api.Group("/users", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	users.Get("/", middleware.RequirePermission(perms, "user:list"), deps.UserService.List)
	users.Post("/", middleware.RequirePermission(perms, "user:update:any"), deps.UserService.Create)
	users.Delete("/:id", middleware.RequirePermission(perms, "user:delete"), deps.UserService.Delete)
	users.Patch("/:id/role", middleware.RequirePermission(perms, "role:assign"), deps.UserService.AssignRole)
	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)
}
