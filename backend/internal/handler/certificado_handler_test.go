package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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

type certFixture struct {
	app        *fiber.App
	db         *gorm.DB
	cursoID    uint
	alunoID    uint
	adminToken string
	profToken  string
	alunoToken string
}

func setupCertFixture(t *testing.T) *certFixture {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))

	mkUser := func(nome, email string, role models.Role, pass string) uint {
		hash, err := auth.HashPassword(pass)
		require.NoError(t, err)
		u := models.User{Nome: nome, Email: email, PasswordHash: hash, Role: role}
		require.NoError(t, db.Create(&u).Error)
		return u.ID
	}
	mkUser("Admin", "admin@test.com", models.RoleAdmin, "admin123")
	mkUser("Professor", "professor@test.com", models.RoleProfessor, "prof123")
	alunoID := mkUser("Aluno Demo", "aluno@test.com", models.RoleAluno, "aluno123")

	cfg := config.Config{
		JWTSecret:     "test-secret",
		JWTExpiration: time.Hour,
		VODBackend:    "local",
		VODLocalDir:   t.TempDir(),
		LiveBackend:   "stub",
	}
	h := handler.New(repository.New(db), cfg)
	tokens := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration)

	app := fiber.New()
	app.Post("/api/v1/auth/login", h.Login)

	protected := app.Group("/api/v1", middleware.Authenticate(tokens))
	requireAdmin := middleware.RequireRoles(models.RoleAdmin)

	protected.Post("/cursos", requireAdmin, h.CreateCurso)
	protected.Get("/cursos/:id/certificado", h.GetCertificadoStatus)
	protected.Get("/cursos/:id/certificado/pdf", h.GetCertificadoPDF)
	protected.Put("/cursos/:id/certificados/elegibilidade", requireAdmin, h.PutCertificadoElegibilidade)
	protected.Get("/cursos/:id/certificados", requireAdmin, h.ListCertificadosCurso)
	protected.Post("/cursos/:id/certificados/:certId/invalidar", requireAdmin, h.InvalidarCertificado)

	login := func(email, pass string) string {
		_, data := jsonRequest(t, app, "POST", "/api/v1/auth/login", map[string]string{
			"email": email, "password": pass,
		}, "")
		var resp struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.Unmarshal(data, &resp))
		require.NotEmpty(t, resp.Token)
		return resp.Token
	}

	fx := &certFixture{
		app:        app,
		db:         db,
		alunoID:    alunoID,
		adminToken: login("admin@test.com", "admin123"),
		profToken:  login("professor@test.com", "prof123"),
		alunoToken: login("aluno@test.com", "aluno123"),
	}

	_, data := jsonRequest(t, app, "POST", "/api/v1/cursos", map[string]string{
		"titulo": "Aula Magna — Direito Constitucional", "descricao": "demo",
	}, fx.adminToken)
	var curso struct {
		ID uint `json:"id"`
	}
	require.NoError(t, json.Unmarshal(data, &curso))
	fx.cursoID = curso.ID
	return fx
}

func certPath(cursoID uint, suffix string) string {
	return "/api/v1/cursos/" + strconv.FormatUint(uint64(cursoID), 10) + suffix
}

func TestCertificadoAdminMarcaElegibilidadeSemEmitir(t *testing.T) {
	fx := setupCertFixture(t)

	resp, data := jsonRequest(t, fx.app, "PUT", certPath(fx.cursoID, "/certificados/elegibilidade"), map[string]any{
		"user_id": fx.alunoID,
	}, fx.adminToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), `"elegivel":true`)

	var count int64
	require.NoError(t, fx.db.Model(&models.Certificado{}).Count(&count).Error)
	require.Equal(t, int64(0), count)

	resp, data = jsonRequest(t, fx.app, "GET", certPath(fx.cursoID, "/certificado"), nil, fx.alunoToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	var status map[string]any
	require.NoError(t, json.Unmarshal(data, &status))
	require.Equal(t, true, status["elegivel"])
	require.Nil(t, status["certificado"])
}

func TestCertificadoLazyEmitEReuso(t *testing.T) {
	fx := setupCertFixture(t)

	resp, data := jsonRequest(t, fx.app, "PUT", certPath(fx.cursoID, "/certificados/elegibilidade"), map[string]any{
		"user_id": fx.alunoID,
	}, fx.adminToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))

	req := httptest.NewRequest(http.MethodGet, certPath(fx.cursoID, "/certificado/pdf"), nil)
	req.Header.Set("Authorization", "Bearer "+fx.alunoToken)
	resp1, err := fx.app.Test(req, -1)
	require.NoError(t, err)
	body1 := mustRead(t, resp1)
	require.Equal(t, fiber.StatusOK, resp1.StatusCode, string(body1))
	require.Equal(t, "application/pdf", resp1.Header.Get("Content-Type"))
	require.True(t, len(body1) > 100)
	require.Equal(t, "%PDF", string(body1[:4]))

	var cert1 models.Certificado
	require.NoError(t, fx.db.Where("status = ?", models.CertificadoStatusValido).First(&cert1).Error)

	req2 := httptest.NewRequest(http.MethodGet, certPath(fx.cursoID, "/certificado/pdf"), nil)
	req2.Header.Set("Authorization", "Bearer "+fx.alunoToken)
	resp2, err := fx.app.Test(req2, -1)
	require.NoError(t, err)
	body2 := mustRead(t, resp2)
	require.Equal(t, fiber.StatusOK, resp2.StatusCode, string(body2))
	require.Equal(t, "application/pdf", resp2.Header.Get("Content-Type"))

	var count int64
	require.NoError(t, fx.db.Model(&models.Certificado{}).Where("status = ?", models.CertificadoStatusValido).Count(&count).Error)
	require.Equal(t, int64(1), count)

	resp, data = jsonRequest(t, fx.app, "GET", certPath(fx.cursoID, "/certificado"), nil, fx.alunoToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), `"id":`+strconv.FormatUint(uint64(cert1.ID), 10))
}

func TestCertificadoInvalidarBloqueiaEReabilitarNovoID(t *testing.T) {
	fx := setupCertFixture(t)

	jsonRequest(t, fx.app, "PUT", certPath(fx.cursoID, "/certificados/elegibilidade"), map[string]any{
		"user_id": fx.alunoID,
	}, fx.adminToken)

	req := httptest.NewRequest(http.MethodGet, certPath(fx.cursoID, "/certificado/pdf"), nil)
	req.Header.Set("Authorization", "Bearer "+fx.alunoToken)
	resp, err := fx.app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var cert1 models.Certificado
	require.NoError(t, fx.db.Where("status = ?", models.CertificadoStatusValido).First(&cert1).Error)

	resp, data := jsonRequest(t, fx.app, "POST",
		certPath(fx.cursoID, "/certificados/"+strconv.FormatUint(uint64(cert1.ID), 10)+"/invalidar"),
		nil, fx.adminToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), `"status":"invalidado"`)

	req2 := httptest.NewRequest(http.MethodGet, certPath(fx.cursoID, "/certificado/pdf"), nil)
	req2.Header.Set("Authorization", "Bearer "+fx.alunoToken)
	resp2, err := fx.app.Test(req2, -1)
	require.NoError(t, err)
	body2 := mustRead(t, resp2)
	require.Equal(t, fiber.StatusForbidden, resp2.StatusCode, string(body2))
	require.Contains(t, string(body2), "não está elegível")

	resp, data = jsonRequest(t, fx.app, "PUT", certPath(fx.cursoID, "/certificados/elegibilidade"), map[string]any{
		"user_id": fx.alunoID,
	}, fx.adminToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))

	req3 := httptest.NewRequest(http.MethodGet, certPath(fx.cursoID, "/certificado/pdf"), nil)
	req3.Header.Set("Authorization", "Bearer "+fx.alunoToken)
	resp3, err := fx.app.Test(req3, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp3.StatusCode)

	var cert2 models.Certificado
	require.NoError(t, fx.db.Where("status = ?", models.CertificadoStatusValido).First(&cert2).Error)
	require.NotEqual(t, cert1.ID, cert2.ID)

	var invalidado models.Certificado
	require.NoError(t, fx.db.First(&invalidado, cert1.ID).Error)
	require.Equal(t, models.CertificadoStatusInvalidado, invalidado.Status)
}

func TestCertificadoRBACGestao(t *testing.T) {
	fx := setupCertFixture(t)
	body := map[string]any{"user_id": fx.alunoID}

	resp, data := jsonRequest(t, fx.app, "PUT", certPath(fx.cursoID, "/certificados/elegibilidade"), body, fx.profToken)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode, string(data))
	require.Contains(t, string(data), "sem permissão")

	resp, data = jsonRequest(t, fx.app, "PUT", certPath(fx.cursoID, "/certificados/elegibilidade"), body, fx.alunoToken)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "GET", certPath(fx.cursoID, "/certificados"), nil, fx.profToken)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "PUT", certPath(fx.cursoID, "/certificados/elegibilidade"), body, "")
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, string(data))
	require.Contains(t, string(data), "não autenticado")
}

func TestCertificadoPDFSemElegibilidade(t *testing.T) {
	fx := setupCertFixture(t)
	req := httptest.NewRequest(http.MethodGet, certPath(fx.cursoID, "/certificado/pdf"), nil)
	req.Header.Set("Authorization", "Bearer "+fx.alunoToken)
	resp, err := fx.app.Test(req, -1)
	require.NoError(t, err)
	body := mustRead(t, resp)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode, string(body))
	require.Contains(t, string(body), "não está elegível")
}

func TestCertificadoElegibilidadeNaoAluno(t *testing.T) {
	fx := setupCertFixture(t)
	var prof models.User
	require.NoError(t, fx.db.Where("email = ?", "professor@test.com").First(&prof).Error)

	resp, data := jsonRequest(t, fx.app, "PUT", certPath(fx.cursoID, "/certificados/elegibilidade"), map[string]any{
		"user_id": prof.ID,
	}, fx.adminToken)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, string(data))
	require.Contains(t, string(data), "não é aluno")
}

func mustRead(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return data
}
