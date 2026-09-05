package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
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
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/vodstorage"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func minimalMP4() []byte {
	return []byte{
		0x00, 0x00, 0x00, 0x18, 'f', 't', 'y', 'p',
		'i', 's', 'o', 'm', 0x00, 0x00, 0x02, 0x00,
		'i', 's', 'o', 'm', 'i', 's', 'o', '2',
		0x00, 0x00, 0x00, 0x08, 'f', 'r', 'e', 'e',
	}
}

func createAulaForVod(t *testing.T, app *fiber.App, adminToken, profToken string) uint {
	t.Helper()
	_, data := jsonRequest(t, app, "POST", "/api/v1/cursos", map[string]string{
		"titulo": "Curso VOD " + t.Name(), "descricao": "demo",
	}, adminToken)
	var curso struct {
		ID uint `json:"id"`
	}
	require.NoError(t, json.Unmarshal(data, &curso))
	resp, data := jsonRequest(t, app, "POST", "/api/v1/aulas", map[string]any{
		"curso_id": curso.ID, "titulo": "Aula", "agendada_em": "04/09/2026 20:00",
	}, profToken)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode, string(data))
	var aula struct {
		ID uint `json:"id"`
	}
	require.NoError(t, json.Unmarshal(data, &aula))
	return aula.ID
}

func setupVodFixture(t *testing.T) (app *fiber.App, aulaID uint, profToken, alunoToken, adminToken, vodDir string) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))

	vodDir = t.TempDir()
	store, err := vodstorage.NewLocal(vodDir)
	require.NoError(t, err)

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
		JWTSecret:      "test-secret",
		JWTExpiration:  time.Hour,
		VODBackend:     "local",
		VODLocalDir:    vodDir,
		VODPlaybackTTL: 15 * time.Minute,
	}
	h := handler.New(repository.New(db), cfg).WithStorage(store)
	tokens := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration)

	app = fiber.New(fiber.Config{BodyLimit: 55 * 1024 * 1024})
	app.Post("/api/v1/auth/login", h.Login)
	app.Get("/api/v1/aulas/:id/vod/content", h.GetVodContent)

	protected := app.Group("/api/v1", middleware.Authenticate(tokens))
	requireAdmin := middleware.RequireRoles(models.RoleAdmin)
	requireAdminProfessor := middleware.RequireRoles(models.RoleAdmin, models.RoleProfessor)

	protected.Post("/cursos", requireAdmin, h.CreateCurso)
	protected.Post("/aulas", requireAdminProfessor, h.CreateAula)
	protected.Delete("/aulas/:id", requireAdminProfessor, h.DeleteAula)
	protected.Get("/aulas/:id/vod", h.GetVod)
	protected.Get("/aulas/:id/vod/playback", h.GetVodPlayback)
	protected.Put("/aulas/:id/vod", requireAdminProfessor, h.PutVod)
	protected.Delete("/aulas/:id/vod", requireAdminProfessor, h.DeleteVod)

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
	adminToken = login("admin@test.com", "admin123")
	profToken = login("professor@test.com", "prof123")
	alunoToken = login("aluno@test.com", "aluno123")
	aulaID = createAulaForVod(t, app, adminToken, profToken)
	return app, aulaID, profToken, alunoToken, adminToken, vodDir
}

func multipartPut(t *testing.T, app *fiber.App, path, field, filename string, content []byte, token string) (*http.Response, []byte) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile(field, filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPut, path, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, data
}

func playbackRequestPath(t *testing.T, playbackURL string) string {
	t.Helper()
	u, err := url.Parse(playbackURL)
	require.NoError(t, err)
	if u.RawQuery != "" {
		return u.Path + "?" + u.RawQuery
	}
	return u.Path
}

func TestPutVodProfessorOKAndGetMetadataPlaybackContent(t *testing.T) {
	app, aulaID, profToken, alunoToken, _, vodDir := setupVodFixture(t)
	path := "/api/v1/aulas/" + strconv.FormatUint(uint64(aulaID), 10) + "/vod"

	resp, data := multipartPut(t, app, path, "file", "aula.mp4", minimalMP4(), profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), `"status":"publicado"`)
	require.Contains(t, string(data), `"content_type":"video/mp4"`)

	keyPath := filepath.Join(vodDir, filepath.FromSlash(vodstorage.ObjectKey(aulaID)))
	_, err := os.Stat(keyPath)
	require.NoError(t, err)

	resp, data = jsonRequest(t, app, "GET", path, nil, alunoToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	require.Contains(t, string(data), "publicado")

	resp, data = jsonRequest(t, app, "GET", path+"/playback", nil, alunoToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, string(data))
	var playback struct {
		PlaybackURL      string `json:"playback_url"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	require.NoError(t, json.Unmarshal(data, &playback))
	require.NotEmpty(t, playback.PlaybackURL)
	require.Equal(t, 900, playback.ExpiresInSeconds)

	req := httptest.NewRequest(http.MethodGet, playbackRequestPath(t, playback.PlaybackURL), nil)
	contentResp, err := app.Test(req, -1)
	require.NoError(t, err)
	content, err := io.ReadAll(contentResp.Body)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, contentResp.StatusCode, string(content))
	require.Equal(t, minimalMP4(), content)
	require.Equal(t, "video/mp4", contentResp.Header.Get("Content-Type"))
}

func TestPutVodRejectsNonMP4(t *testing.T) {
	app, aulaID, profToken, _, _, _ := setupVodFixture(t)
	path := "/api/v1/aulas/" + strconv.FormatUint(uint64(aulaID), 10) + "/vod"
	resp, data := multipartPut(t, app, path, "file", "aula.txt", []byte("não é mp4"), profToken)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, string(data))
	require.Contains(t, string(data), "MP4")
}

func TestPutVodRejectsOversize(t *testing.T) {
	// Limite de 50 MB é validado em vodstorage.ValidateUpload (header.Size) antes do Put.
	// Cobertura HTTP de 50MB+ seria lenta em CI/Docker volume; ver validate_test.go.
	h := &multipart.FileHeader{
		Filename: "aula.mp4",
		Size:     models.VodMaxSizeBytes + 1,
		Header:   textproto.MIMEHeader{"Content-Type": []string{"video/mp4"}},
	}
	err := vodstorage.ValidateUpload(h, func() (multipart.File, error) {
		t.Fatal("não deve abrir")
		return nil, nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "50 MB")
}

func TestAlunoCannotPutOrDeleteVod(t *testing.T) {
	app, aulaID, profToken, alunoToken, _, _ := setupVodFixture(t)
	path := "/api/v1/aulas/" + strconv.FormatUint(uint64(aulaID), 10) + "/vod"

	resp, data := multipartPut(t, app, path, "file", "aula.mp4", minimalMP4(), alunoToken)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode, string(data))
	require.Contains(t, string(data), "sem permissão")

	resp, _ = multipartPut(t, app, path, "file", "aula.mp4", minimalMP4(), profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	req := httptest.NewRequest(http.MethodDelete, path, nil)
	req.Header.Set("Authorization", "Bearer "+alunoToken)
	delResp, err := app.Test(req, -1)
	require.NoError(t, err)
	body, err := io.ReadAll(delResp.Body)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusForbidden, delResp.StatusCode, string(body))
}

func TestDeleteVodRemovesObject(t *testing.T) {
	app, aulaID, profToken, _, _, vodDir := setupVodFixture(t)
	path := "/api/v1/aulas/" + strconv.FormatUint(uint64(aulaID), 10) + "/vod"
	resp, _ := multipartPut(t, app, path, "file", "aula.mp4", minimalMP4(), profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	keyPath := filepath.Join(vodDir, filepath.FromSlash(vodstorage.ObjectKey(aulaID)))
	_, err := os.Stat(keyPath)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, path, nil)
	req.Header.Set("Authorization", "Bearer "+profToken)
	delResp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, delResp.StatusCode)

	_, err = os.Stat(keyPath)
	require.True(t, os.IsNotExist(err))

	getResp, data := jsonRequest(t, app, "GET", path, nil, profToken)
	require.Equal(t, fiber.StatusNotFound, getResp.StatusCode)
	require.Contains(t, string(data), "Gravação não encontrada")
}

func TestDeleteAulaCascadesVod(t *testing.T) {
	app, aulaID, profToken, _, _, vodDir := setupVodFixture(t)
	path := "/api/v1/aulas/" + strconv.FormatUint(uint64(aulaID), 10) + "/vod"
	resp, _ := multipartPut(t, app, path, "file", "aula.mp4", minimalMP4(), profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	keyPath := filepath.Join(vodDir, filepath.FromSlash(vodstorage.ObjectKey(aulaID)))
	_, err := os.Stat(keyPath)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/aulas/"+strconv.FormatUint(uint64(aulaID), 10), nil)
	req.Header.Set("Authorization", "Bearer "+profToken)
	delResp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNoContent, delResp.StatusCode)

	_, err = os.Stat(keyPath)
	require.True(t, os.IsNotExist(err))
}

func TestVodContentRejectsInvalidToken(t *testing.T) {
	app, aulaID, profToken, _, _, _ := setupVodFixture(t)
	path := "/api/v1/aulas/" + strconv.FormatUint(uint64(aulaID), 10) + "/vod"
	resp, _ := multipartPut(t, app, path, "file", "aula.mp4", minimalMP4(), profToken)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	req := httptest.NewRequest(http.MethodGet, path+"/content?token=invalido", nil)
	contentResp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, contentResp.StatusCode)
}
