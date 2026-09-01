package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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

func setupTestApp(t *testing.T) (*fiber.App, string) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))

	hash, err := auth.HashPassword("admin123")
	require.NoError(t, err)
	require.NoError(t, db.Create(&models.User{
		Nome:         "Admin",
		Email:        "admin@test.com",
		PasswordHash: hash,
		Role:         models.RoleAdmin,
	}).Error)

	cfg := config.Config{JWTSecret: "test-secret", JWTExpiration: time.Hour}
	h := handler.New(repository.New(db), cfg)
	tokens := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration)

	app := fiber.New()
	app.Post("/api/v1/auth/login", h.Login)

	protected := app.Group("/api/v1", middleware.Authenticate(tokens))
	protected.Get("/auth/me", h.Me)
	protected.Get("/cursos", h.ListCursos)
	protected.Post("/cursos", h.CreateCurso)
	protected.Get("/aulas", h.ListAulas)
	protected.Post("/aulas", h.CreateAula)
	protected.Get("/alunos", h.ListAlunos)
	protected.Post("/alunos", h.CreateAluno)

	_, data := jsonRequest(t, app, "POST", "/api/v1/auth/login", map[string]string{
		"email":    "admin@test.com",
		"password": "admin123",
	}, "")
	var loginResp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(data, &loginResp))
	require.NotEmpty(t, loginResp.Token)

	return app, loginResp.Token
}

func jsonRequest(t *testing.T, app *fiber.App, method, path string, body any, token string) (*http.Response, []byte) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, data
}

func TestLogin(t *testing.T) {
	app, token := setupTestApp(t)
	resp, data := jsonRequest(t, app, "GET", "/api/v1/auth/me", nil, token)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Contains(t, string(data), "admin@test.com")
}

func TestCursoCRUD(t *testing.T) {
	app, token := setupTestApp(t)

	resp, _ := jsonRequest(t, app, "POST", "/api/v1/cursos", map[string]string{
		"titulo":    "Direito Constitucional",
		"descricao": "Aula magna",
	}, token)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	resp, data := jsonRequest(t, app, "GET", "/api/v1/cursos", nil, token)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Contains(t, string(data), "Direito Constitucional")
}

func TestAulaRequiresCurso(t *testing.T) {
	app, token := setupTestApp(t)
	resp, data := jsonRequest(t, app, "POST", "/api/v1/aulas", map[string]any{
		"curso_id":    999,
		"titulo":      "Aula 1",
		"agendada_em": "29/08/2026 19:00",
	}, token)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.Contains(t, string(data), "curso informado não existe")
}

func TestAlunoDuplicateEmail(t *testing.T) {
	app, token := setupTestApp(t)
	payload := map[string]string{"nome": "Maria", "email": "maria@test.com"}
	resp, _ := jsonRequest(t, app, "POST", "/api/v1/alunos", payload, token)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	resp, data := jsonRequest(t, app, "POST", "/api/v1/alunos", payload, token)
	require.Equal(t, fiber.StatusConflict, resp.StatusCode)
	require.Contains(t, string(data), "duplicado")
}

func TestUnauthorizedWithoutToken(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))

	cfg := config.Config{JWTSecret: "test-secret", JWTExpiration: time.Hour}
	h := handler.New(repository.New(db), cfg)
	tokens := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration)

	app := fiber.New()
	protected := app.Group("/api/v1", middleware.Authenticate(tokens))
	protected.Get("/cursos", h.ListCursos)

	resp, _ := jsonRequest(t, app, "GET", "/api/v1/cursos", nil, "")
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
