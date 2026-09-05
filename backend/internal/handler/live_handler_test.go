package handler_test

import (
	"encoding/json"
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

type liveFixture struct {
	app        *fiber.App
	db         *gorm.DB
	aulaID     uint
	aula2ID    uint
	profToken  string
	alunoToken string
	adminToken string
}

func setupLiveFixture(t *testing.T, liveBackend string) *liveFixture {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))

	mkUser := func(nome, email string, role models.Role, pass string) {
		hash, err := auth.HashPassword(pass)
		require.NoError(t, err)
		require.NoError(t, db.Create(&models.User{
			Nome: nome, Email: email, PasswordHash: hash, Role: role,
		}).Error)
	}
	mkUser("Admin", "admin@test.com", models.RoleAdmin, "admin123")
	mkUser("Professor", "professor@test.com", models.RoleProfessor, "prof123")
	mkUser("Aluno", "aluno@test.com", models.RoleAluno, "aluno123")

	cfg := config.Config{
		JWTSecret:     "test-secret",
		JWTExpiration: time.Hour,
		LiveBackend:   liveBackend,
		VODBackend:    "local",
		VODLocalDir:   t.TempDir(),
	}
	h := handler.New(repository.New(db), cfg)
	tokens := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration)

	app := fiber.New()
	app.Post("/api/v1/auth/login", h.Login)

	protected := app.Group("/api/v1", middleware.Authenticate(tokens))
	requireAdmin := middleware.RequireRoles(models.RoleAdmin)
	requireAdminProfessor := middleware.RequireRoles(models.RoleAdmin, models.RoleProfessor)

	protected.Post("/cursos", requireAdmin, h.CreateCurso)
	protected.Post("/aulas", requireAdminProfessor, h.CreateAula)
	protected.Get("/aulas/:id", h.GetAula)
	protected.Delete("/aulas/:id", requireAdminProfessor, h.DeleteAula)

	protected.Get("/aulas/:id/live", h.GetLive)
	protected.Post("/aulas/:id/live/schedule", requireAdminProfessor, h.ScheduleLive)
	protected.Post("/aulas/:id/live/cancel", requireAdminProfessor, h.CancelLive)
	protected.Post("/aulas/:id/live/start", requireAdminProfessor, h.StartLive)
	protected.Post("/aulas/:id/live/stop", requireAdminProfessor, h.StopLive)
	protected.Get("/aulas/:id/live/playback", h.GetLivePlayback)
	protected.Get("/aulas/:id/live/ingest", requireAdminProfessor, h.GetLiveIngest)

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

	fx := &liveFixture{
		app:        app,
		db:         db,
		adminToken: login("admin@test.com", "admin123"),
		profToken:  login("professor@test.com", "prof123"),
		alunoToken: login("aluno@test.com", "aluno123"),
	}

	_, data := jsonRequest(t, app, "POST", "/api/v1/cursos", map[string]string{
		"titulo": "Curso Live", "descricao": "demo",
	}, fx.adminToken)
	var curso struct {
		ID uint `json:"id"`
	}
	require.NoError(t, json.Unmarshal(data, &curso))

	createAula := func(titulo string) uint {
		resp, data := jsonRequest(t, app, "POST", "/api/v1/aulas", map[string]any{
			"curso_id": curso.ID, "titulo": titulo, "agendada_em": "05/09/2026 20:00",
		}, fx.profToken)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode, string(data))
		var aula struct {
			ID uint `json:"id"`
		}
		require.NoError(t, json.Unmarshal(data, &aula))
		return aula.ID
	}
	fx.aulaID = createAula("Aula 1")
	fx.aula2ID = createAula("Aula 2")
	return fx
}

func livePath(aulaID uint, suffix string) string {
	return "/api/v1/aulas/" + strconv.FormatUint(uint64(aulaID), 10) + "/live" + suffix
}

func TestGetLiveInativa(t *testing.T) {
	fx := setupLiveFixture(t, "stub")
	resp, data := jsonRequest(t, fx.app, "GET", livePath(fx.aulaID, ""), nil, fx.alunoToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	var body map[string]any
	require.NoError(t, json.Unmarshal(data, &body))
	require.Equal(t, "inativa", body["status"])
	require.Equal(t, "stub", body["modo"])
	require.Nil(t, body["iniciada_em"])
	_, hasKey := body["stream_key"]
	require.False(t, hasKey)
}

func TestLiveProfessorStartStopRestart(t *testing.T) {
	fx := setupLiveFixture(t, "stub")

	resp, data := jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	var started map[string]any
	require.NoError(t, json.Unmarshal(data, &started))
	require.Equal(t, "ao_vivo", started["status"])
	require.NotNil(t, started["iniciada_em"])

	// Idempotente nesta aula
	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))

	// Aula.Status CRUD inalterado
	resp, data = jsonRequest(t, fx.app, "GET", "/api/v1/aulas/"+strconv.FormatUint(uint64(fx.aulaID), 10), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	var aula models.Aula
	require.NoError(t, json.Unmarshal(data, &aula))
	require.Equal(t, models.AulaStatusAgendada, aula.Status)

	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/stop"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	var stopped map[string]any
	require.NoError(t, json.Unmarshal(data, &stopped))
	require.Equal(t, "encerrada", stopped["status"])

	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), `"status":"ao_vivo"`)
}

func TestLiveStartSemAgendadaEm(t *testing.T) {
	fx := setupLiveFixture(t, "stub")
	aula := models.Aula{
		CursoID: 1,
		Titulo:  "Sem horário",
		Status:  models.AulaStatusAgendada,
	}
	require.NoError(t, fx.db.Create(&aula).Error)

	resp, data := jsonRequest(t, fx.app, "POST", livePath(aula.ID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), `"status":"ao_vivo"`)
}

func TestLiveAlunoForbiddenEscritaEIngest(t *testing.T) {
	fx := setupLiveFixture(t, "stub")

	resp, data := jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.alunoToken)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/stop"), nil, fx.alunoToken)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "GET", livePath(fx.aulaID, "/ingest"), nil, fx.alunoToken)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode, string(data))
	require.NotContains(t, string(data), "stream_key")
}

func TestLiveConflictSegundaAulaAoVivo(t *testing.T) {
	fx := setupLiveFixture(t, "stub")

	resp, data := jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aula2ID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusConflict, resp.StatusCode, string(data))
	require.Contains(t, string(data), "Já existe uma transmissão ao vivo")
}

func TestLivePlaybackStubSemURL(t *testing.T) {
	fx := setupLiveFixture(t, "stub")

	resp, data := jsonRequest(t, fx.app, "GET", livePath(fx.aulaID, "/playback"), nil, fx.alunoToken)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "GET", livePath(fx.aulaID, "/playback"), nil, fx.alunoToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	var body map[string]any
	require.NoError(t, json.Unmarshal(data, &body))
	require.Equal(t, "stub", body["modo"])
	require.Nil(t, body["playback_url"])
	_, hasKey := body["stream_key"]
	require.False(t, hasKey)
}

func TestLiveAnonimoUnauthorized(t *testing.T) {
	fx := setupLiveFixture(t, "stub")
	for _, tc := range []struct {
		method string
		suffix string
	}{
		{"GET", ""},
		{"POST", "/schedule"},
		{"POST", "/cancel"},
		{"POST", "/start"},
		{"POST", "/stop"},
		{"GET", "/playback"},
		{"GET", "/ingest"},
	} {
		path := livePath(fx.aulaID, tc.suffix)
		resp, data := jsonRequest(t, fx.app, tc.method, path, nil, "")
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, path+" "+string(data))
	}
}

func TestLiveStopSemAoVivo(t *testing.T) {
	fx := setupLiveFixture(t, "stub")
	resp, data := jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/stop"), nil, fx.profToken)
	require.Equal(t, fiber.StatusConflict, resp.StatusCode, string(data))
}

func TestLiveIngestStub(t *testing.T) {
	fx := setupLiveFixture(t, "stub")
	resp, data := jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "GET", livePath(fx.aulaID, "/ingest"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	var body map[string]any
	require.NoError(t, json.Unmarshal(data, &body))
	require.Equal(t, "stub", body["modo"])
	require.Nil(t, body["ingest_server"])
	require.Nil(t, body["stream_key"])
}

func TestLiveStartIVSSemCredenciais503(t *testing.T) {
	fx := setupLiveFixture(t, "ivs")
	resp, data := jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "GET", livePath(fx.aulaID, ""), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), `"status":"inativa"`)
}

func TestDeleteAulaCascadesLive(t *testing.T) {
	fx := setupLiveFixture(t, "stub")
	resp, data := jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "DELETE", "/api/v1/aulas/"+strconv.FormatUint(uint64(fx.aulaID), 10), nil, fx.profToken)
	require.Equal(t, fiber.StatusNoContent, resp.StatusCode, string(data))

	var count int64
	require.NoError(t, fx.db.Model(&models.AulaLive{}).Where("aula_id = ?", fx.aulaID).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestLiveScheduleRequiresHorario(t *testing.T) {
	fx := setupLiveFixture(t, "stub")
	aula := models.Aula{
		CursoID: 1,
		Titulo:  "Sem horário para agendar",
		Status:  models.AulaStatusAgendada,
	}
	require.NoError(t, fx.db.Create(&aula).Error)

	resp, data := jsonRequest(t, fx.app, "POST", livePath(aula.ID, "/schedule"), nil, fx.profToken)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, string(data))
	require.Contains(t, string(data), "agendada_em")
}

func TestLiveScheduleStartCancel(t *testing.T) {
	fx := setupLiveFixture(t, "stub")

	resp, data := jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/schedule"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	var scheduled map[string]any
	require.NoError(t, json.Unmarshal(data, &scheduled))
	require.Equal(t, "agendada", scheduled["status"])
	require.Nil(t, scheduled["iniciada_em"])

	// Idempotente
	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/schedule"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))

	// Playback não disponível em agendada
	resp, data = jsonRequest(t, fx.app, "GET", livePath(fx.aulaID, "/playback"), nil, fx.alunoToken)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode, string(data))

	// agendada → ao_vivo
	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), `"status":"ao_vivo"`)

	// ao_vivo → encerrada
	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/stop"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), `"status":"encerrada"`)

	// Reagendar após encerrada
	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/schedule"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), `"status":"agendada"`)

	// Cancelar → inativa
	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/cancel"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), `"status":"inativa"`)

	var count int64
	require.NoError(t, fx.db.Model(&models.AulaLive{}).Where("aula_id = ?", fx.aulaID).Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestLiveCancelSemAgendada(t *testing.T) {
	fx := setupLiveFixture(t, "stub")
	resp, data := jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/cancel"), nil, fx.profToken)
	require.Equal(t, fiber.StatusConflict, resp.StatusCode, string(data))
}

func TestLiveAlunoForbiddenScheduleCancel(t *testing.T) {
	fx := setupLiveFixture(t, "stub")
	resp, data := jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/schedule"), nil, fx.alunoToken)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/cancel"), nil, fx.alunoToken)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode, string(data))
}

func TestLiveScheduleEnquantoAoVivo(t *testing.T) {
	fx := setupLiveFixture(t, "stub")
	resp, data := jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/start"), nil, fx.profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))

	resp, data = jsonRequest(t, fx.app, "POST", livePath(fx.aulaID, "/schedule"), nil, fx.profToken)
	require.Equal(t, fiber.StatusConflict, resp.StatusCode, string(data))
}
