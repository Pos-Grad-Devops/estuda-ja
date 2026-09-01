package main

import (
	"log"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/auth"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/database"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/handler"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/middleware"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	if err := database.RunMigrations(db, cfg); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	repos := repository.New(db)
	handlers := handler.New(repos, cfg)
	tokens := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration)

	app := fiber.New(fiber.Config{
		AppName: "EstudaJá API",
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigin,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Get("/health", handlers.Health)

	api := app.Group("/api/v1")
	api.Post("/auth/login", handlers.Login)

	protected := api.Group("", middleware.Authenticate(tokens))
	protected.Get("/auth/me", handlers.Me)

	requireAdmin := middleware.RequireRoles(models.RoleAdmin)
	requireAdminProfessor := middleware.RequireRoles(models.RoleAdmin, models.RoleProfessor)

	protected.Get("/cursos", handlers.ListCursos)
	protected.Get("/cursos/:id", handlers.GetCurso)
	protected.Post("/cursos", requireAdmin, handlers.CreateCurso)
	protected.Put("/cursos/:id", requireAdmin, handlers.UpdateCurso)
	protected.Delete("/cursos/:id", requireAdmin, handlers.DeleteCurso)

	protected.Get("/aulas", handlers.ListAulas)
	protected.Get("/aulas/:id", handlers.GetAula)
	protected.Post("/aulas", requireAdminProfessor, handlers.CreateAula)
	protected.Put("/aulas/:id", requireAdminProfessor, handlers.UpdateAula)
	protected.Delete("/aulas/:id", requireAdminProfessor, handlers.DeleteAula)

	protected.Get("/alunos", requireAdmin, handlers.ListAlunos)
	protected.Post("/alunos", requireAdmin, handlers.CreateAluno)
	protected.Get("/alunos/:id", requireAdmin, handlers.GetAluno)
	protected.Put("/alunos/:id", requireAdmin, handlers.UpdateAluno)
	protected.Delete("/alunos/:id", requireAdmin, handlers.DeleteAluno)

	protected.Get("/users", requireAdmin, handlers.ListUsers)
	protected.Post("/users", requireAdmin, handlers.CreateUser)
	protected.Get("/users/:id", requireAdmin, handlers.GetUser)
	protected.Put("/users/:id", requireAdmin, handlers.UpdateUser)
	protected.Delete("/users/:id", requireAdmin, handlers.DeleteUser)

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}
