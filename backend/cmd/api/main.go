package main

import (
	"log"
	"strings"

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

const vodBodyLimit = 52*1024*1024 + 1024*1024 // 50 MiB + overhead multipart

func main() {
	cfg := config.Load()

	if !config.LiveBackendValid(cfg.LiveBackend) {
		log.Fatalf("LIVE_BACKEND inválido %q (use stub ou ivs)", cfg.LiveBackend)
	}
	if strings.EqualFold(cfg.LiveBackend, "ivs") && !cfg.IVSConfigured() {
		log.Fatal("LIVE_BACKEND=ivs exige IVS_INGEST_ENDPOINT, IVS_STREAM_KEY e IVS_PLAYBACK_URL")
	}
	if strings.EqualFold(cfg.VODBackend, "s3") && strings.TrimSpace(cfg.VODS3Bucket) == "" {
		log.Fatal("VOD_BACKEND=s3 exige VOD_S3_BUCKET")
	}

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
		AppName:   "EstudaJá API",
		BodyLimit: vodBodyLimit,
	})

	app.Use(recover.New())
	// ${path} sem query — não registrar ?token= do handshake WS.
	app.Use(logger.New(logger.Config{
		Format: "${time} | ${status} | ${latency} | ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigin, // origem única via CORS_ORIGIN (sem lista hardcoded)
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Get("/health", handlers.Health)

	api := app.Group("/api/v1")
	api.Post("/auth/login", handlers.Login)

	// Content assinado por query token (sem Bearer) — permite <video src="...">.
	api.Get("/aulas/:id/vod/content", handlers.GetVodContent)

	// Chat WS: JWT em ?token= (sem Bearer). Hub em memória; independente de live/VOD.
	api.Get("/aulas/:id/chat/ws", handlers.ChatHandshake, handlers.ChatWS())

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

	protected.Get("/aulas/:id/vod", handlers.GetVod)
	protected.Get("/aulas/:id/vod/playback", handlers.GetVodPlayback)
	protected.Put("/aulas/:id/vod", requireAdminProfessor, handlers.PutVod)
	protected.Delete("/aulas/:id/vod", requireAdminProfessor, handlers.DeleteVod)

	protected.Get("/aulas/:id/live", handlers.GetLive)
	protected.Post("/aulas/:id/live/schedule", requireAdminProfessor, handlers.ScheduleLive)
	protected.Post("/aulas/:id/live/cancel", requireAdminProfessor, handlers.CancelLive)
	protected.Post("/aulas/:id/live/start", requireAdminProfessor, handlers.StartLive)
	protected.Post("/aulas/:id/live/stop", requireAdminProfessor, handlers.StopLive)
	protected.Get("/aulas/:id/live/playback", handlers.GetLivePlayback)
	protected.Get("/aulas/:id/live/ingest", requireAdminProfessor, handlers.GetLiveIngest)

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
	log.Printf("listening on %s (VOD_BACKEND=%s LIVE_BACKEND=%s)", addr, cfg.VODBackend, cfg.LiveBackend)
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}
