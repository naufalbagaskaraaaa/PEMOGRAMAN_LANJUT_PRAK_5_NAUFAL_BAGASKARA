package main

import (
	"context"
	"fmt"
	"log"

	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/config"
	"latihan-fiber/database"
	appMiddleware "latihan-fiber/middleware"
	"latihan-fiber/route"

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
	studentService := service.NewStudentService(studentRepo)

	app := fiber.New(fiber.Config{
		AppName:      "Tugas Mandiri REST API - Student Management",
		ErrorHandler: appMiddleware.ErrorHandler,
	})

	appMiddleware.Register(app)

	route.Register(app, pool, studentService)

	app.Use(appMiddleware.NotFound)

	fmt.Println("Server running at http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
