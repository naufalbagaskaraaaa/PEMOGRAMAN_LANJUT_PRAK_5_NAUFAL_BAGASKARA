package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/config"
	"latihan-fiber/database"
	"latihan-fiber/helper"
	appMiddleware "latihan-fiber/middleware"
	"latihan-fiber/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatalf("memuat .env: %v", err)
	}

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	studentRepo := repository.NewStudentRepository(pool)
	prestasiRepo := repository.NewPrestasiRepository(pool)

	authRepo := repository.NewAuthRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	roleRepo := repository.NewRoleRepository(pool)

	rawPermissions, err := roleRepo.LoadPermissions(ctx)
	if err != nil {
		log.Fatalf("gagal memuat permission: %v", err)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	log.Printf("permission dimuat untuk roles: %v", permissions.KnownRoles())

	studentService := service.NewStudentService(studentRepo, permissions)
	prestasiService := service.NewPrestasiService(prestasiRepo)

	userService := service.NewUserService(userRepo, permissions)
	jwtSecret, err := config.RequiredEnv("JWT_SECRET")
	if err != nil {
		log.Fatal(err)
	}
	jwtManager := helper.NewJWTManager(jwtSecret, 15*time.Minute)
	authService := service.NewAuthService(authRepo, jwtManager, permissions, 15*time.Minute, 30*24*time.Hour)

	app := fiber.New(fiber.Config{
		AppName:      "Tugas Mandiri REST API - Student Management",
		ErrorHandler: appMiddleware.ErrorHandler,
	})

	appMiddleware.Register(app)

	routes.Register(app, routes.Dependencies{
		Pool:            pool,
		StudentService:  studentService,
		PrestasiService: prestasiService,
		AuthService:     authService,
		UserService:     userService,
		JWT:             jwtManager,
		Permissions:     permissions,
	})

	app.Use(appMiddleware.NotFound)

	fmt.Println("Server running at http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
