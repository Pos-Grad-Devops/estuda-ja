package handler_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/auth"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/database"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/handler"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/middleware"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoleTestApp(t *testing.T, role models.Role) (*fiber.App, string) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))

	hash, err := auth.HashPassword("secret")
	require.NoError(t, err)
	require.NoError(t, db.Create(&models.User{
		Nome:         "Usuário",
		Email:        string(role) + "@test.com",
		PasswordHash: hash,
		Role:         role,
	}).Error)

	cfg := config.Config{JWTSecret: "test-secret", JWTExpiration: time.Hour}
	h := handler.New(repository.New(db), cfg)
	tokens := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration)

	app := fiber.New()
	app.Post("/api/v1/auth/login", h.Login)

	protected := app.Group("/api/v1", middleware.Authenticate(tokens))
	requireAdmin := middleware.RequireRoles(models.RoleAdmin)
	requireAdminProfessor := middleware.RequireRoles(models.RoleAdmin, models.RoleProfessor)

	protected.Get("/cursos", h.ListCursos)
	protected.Get("/aulas", h.ListAulas)
	protected.Post("/cursos", requireAdmin, h.CreateCurso)
	protected.Post("/aulas", requireAdminProfessor, h.CreateAula)
	protected.Get("/alunos", requireAdmin, h.ListAlunos)

	_, data := jsonRequest(t, app, "POST", "/api/v1/auth/login", map[string]string{
		"email":    string(role) + "@test.com",
		"password": "secret",
	}, "")
	var loginResp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(data, &loginResp))
	return app, loginResp.Token
}

func TestAlunoCannotCreateCurso(t *testing.T) {
	app, token := setupRoleTestApp(t, models.RoleAluno)
	resp, _ := jsonRequest(t, app, "POST", "/api/v1/cursos", map[string]string{
		"titulo": "Curso",
	}, token)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestProfessorCanCreateAula(t *testing.T) {
	app, token := setupRoleTestApp(t, models.RoleProfessor)

	resp, _ := jsonRequest(t, app, "POST", "/api/v1/cursos", map[string]string{
		"titulo": "Curso",
	}, token)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestAlunoCannotListAlunos(t *testing.T) {
	app, token := setupRoleTestApp(t, models.RoleAluno)
	resp, _ := jsonRequest(t, app, "GET", "/api/v1/alunos", nil, token)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestProfessorCanListCursosAndAulas(t *testing.T) {
	app, token := setupRoleTestApp(t, models.RoleProfessor)

	for _, path := range []string{"/api/v1/cursos", "/api/v1/aulas"} {
		resp, _ := jsonRequest(t, app, "GET", path, nil, token)
		require.Equal(t, fiber.StatusOK, resp.StatusCode, path)
	}
}
