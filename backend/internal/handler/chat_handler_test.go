package handler_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/auth"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/database"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/handler"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/repository"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type chatFixture struct {
	host       string
	aulaID     uint
	aula2ID    uint
	alunoToken string
	profToken  string
	adminToken string
}

type chatFrame struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	AulaID    uint   `json:"aula_id"`
	Texto     string `json:"texto"`
	EnviadoEm string `json:"enviado_em"`
	Error     string `json:"error"`
	Autor     struct {
		ID   uint   `json:"id"`
		Nome string `json:"nome"`
	} `json:"autor"`
}

func setupChatFixture(t *testing.T) *chatFixture {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.Migrate(db))

	mkUser := func(nome, email string, role models.Role, pass string) *models.User {
		hash, err := auth.HashPassword(pass)
		require.NoError(t, err)
		u := &models.User{Nome: nome, Email: email, PasswordHash: hash, Role: role}
		require.NoError(t, db.Create(u).Error)
		return u
	}
	admin := mkUser("Admin", "admin@test.com", models.RoleAdmin, "admin123")
	prof := mkUser("Professor", "professor@test.com", models.RoleProfessor, "prof123")
	aluno := mkUser("Maria Silva", "aluno@test.com", models.RoleAluno, "aluno123")

	curso := models.Curso{Titulo: "Curso Chat", Descricao: "demo"}
	require.NoError(t, db.Create(&curso).Error)
	agendada := timeutil.NewDateTime(time.Date(2026, 9, 5, 20, 0, 0, 0, time.Local))
	aula1 := models.Aula{CursoID: curso.ID, Titulo: "Aula Chat 1", AgendadaEm: agendada, Status: models.AulaStatusAgendada}
	aula2 := models.Aula{CursoID: curso.ID, Titulo: "Aula Chat 2", AgendadaEm: agendada, Status: models.AulaStatusAgendada}
	require.NoError(t, db.Create(&aula1).Error)
	require.NoError(t, db.Create(&aula2).Error)

	cfg := config.Config{
		JWTSecret:     "test-secret",
		JWTExpiration: time.Hour,
		VODBackend:    "local",
		VODLocalDir:   t.TempDir(),
		LiveBackend:   "stub",
	}
	h := handler.New(repository.New(db), cfg)
	tokens := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration)

	sign := func(u *models.User) string {
		tok, err := tokens.Generate(u)
		require.NoError(t, err)
		return tok
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/v1/aulas/:id/chat/ws", h.ChatHandshake, h.ChatWS())

	return &chatFixture{
		host:       listenChatApp(t, app),
		aulaID:     aula1.ID,
		aula2ID:    aula2.ID,
		adminToken: sign(admin),
		profToken:  sign(prof),
		alunoToken: sign(aluno),
	}
}

func listenChatApp(t *testing.T, app *fiber.App) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	go func() {
		_ = app.Listener(ln)
	}()
	t.Cleanup(func() {
		_ = app.Shutdown()
	})
	require.Eventually(t, func() bool {
		c, err := net.Dial("tcp", addr)
		if err != nil {
			return false
		}
		_ = c.Close()
		return true
	}, 2*time.Second, 10*time.Millisecond)
	return addr
}

func chatHTTPURL(host string, aulaID uint, token string) string {
	u := fmt.Sprintf("http://%s/api/v1/aulas/%d/chat/ws", host, aulaID)
	if token != "" {
		u += "?token=" + url.QueryEscape(token)
	}
	return u
}

func chatHandshakeGET(t *testing.T, host string, aulaID uint, token string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, chatHTTPURL(host, aulaID, token), nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, data
}

func chatWSURL(host string, aulaID uint, token string) string {
	return fmt.Sprintf("ws://%s/api/v1/aulas/%d/chat/ws?token=%s", host, aulaID, url.QueryEscape(token))
}

func dialChatWS(t *testing.T, host string, aulaID uint, token string) *websocket.Conn {
	t.Helper()
	u := chatWSURL(host, aulaID, token)
	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	var conn *websocket.Conn
	var lastErr error
	var lastStatus int
	var lastBody string
	require.Eventually(t, func() bool {
		var resp *http.Response
		conn, resp, lastErr = dialer.Dial(u, nil)
		lastStatus = 0
		lastBody = ""
		if resp != nil {
			lastStatus = resp.StatusCode
			if resp.Body != nil {
				b, _ := io.ReadAll(resp.Body)
				lastBody = string(b)
				_ = resp.Body.Close()
			}
		}
		return lastErr == nil
	}, 5*time.Second, 20*time.Millisecond, "ws handshake status=%d body=%s", lastStatus, lastBody)
	require.NoError(t, lastErr, "ws handshake status=%d body=%s", lastStatus, lastBody)
	t.Cleanup(func() { _ = conn.Close() })
	// fasthttp devolve 101 antes do Hijack rodar HandleChatConn; sem isto o
	// cliente pode enviar chat.send antes do Register no hub (flake no CI).
	waitChatReady(t, conn)
	return conn
}

func waitChatReady(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	writeChatJSON(t, conn, map[string]string{"type": "chat.ping"})
	frame := readChatFrame(t, conn)
	require.Equal(t, "chat.pong", frame.Type, "conexão WS ainda não pronta no hub")
}

func readChatFrame(t *testing.T, conn *websocket.Conn) chatFrame {
	t.Helper()
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	_, data, err := conn.ReadMessage()
	require.NoError(t, err, "lendo frame WS")
	var frame chatFrame
	require.NoError(t, json.Unmarshal(data, &frame), string(data))
	return frame
}

func writeChatJSON(t *testing.T, conn *websocket.Conn, v any) {
	t.Helper()
	require.NoError(t, conn.WriteJSON(v))
}

func TestChatHandshakeSemToken(t *testing.T) {
	fx := setupChatFixture(t)
	status, data := chatHandshakeGET(t, fx.host, fx.aulaID, "")
	require.Equal(t, fiber.StatusUnauthorized, status)
	require.Contains(t, string(data), "não autenticado")
}

func TestChatHandshakeTokenInvalido(t *testing.T) {
	fx := setupChatFixture(t)
	status, data := chatHandshakeGET(t, fx.host, fx.aulaID, "token-invalido")
	require.Equal(t, fiber.StatusUnauthorized, status)
	require.Contains(t, string(data), "token inválido")
}

func TestChatHandshakeAulaInexistente(t *testing.T) {
	fx := setupChatFixture(t)
	status, data := chatHandshakeGET(t, fx.host, 9999, fx.alunoToken)
	require.Equal(t, fiber.StatusNotFound, status)
	require.Contains(t, string(data), "aula não encontrada")
}

func TestChatHandshakeAnonimoRejeitado(t *testing.T) {
	fx := setupChatFixture(t)
	u := fmt.Sprintf("ws://%s/api/v1/aulas/%d/chat/ws", fx.host, fx.aulaID)
	dialer := websocket.Dialer{HandshakeTimeout: 2 * time.Second}
	_, resp, err := dialer.Dial(u, nil)
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Contains(t, string(body), "não autenticado")
}

func TestChatSendValidoBroadcastEEco(t *testing.T) {
	fx := setupChatFixture(t)
	aluno := dialChatWS(t, fx.host, fx.aulaID, fx.alunoToken)
	prof := dialChatWS(t, fx.host, fx.aulaID, fx.profToken)

	writeChatJSON(t, aluno, map[string]string{"type": "chat.send", "texto": "Bom dia"})

	gotAluno := readChatFrame(t, aluno)
	gotProf := readChatFrame(t, prof)
	require.Equal(t, "chat.message", gotAluno.Type)
	require.Equal(t, "chat.message", gotProf.Type)
	require.Equal(t, fx.aulaID, gotAluno.AulaID)
	require.Equal(t, "Bom dia", gotAluno.Texto)
	require.Equal(t, "Maria Silva", gotAluno.Autor.Nome)
	require.Equal(t, gotAluno.ID, gotProf.ID)
	require.NotEmpty(t, gotAluno.EnviadoEm)
	require.Regexp(t, `\d{2}/\d{2}/\d{4} \d{2}:\d{2}`, gotAluno.EnviadoEm)
}

func TestChatSendVazioEAcimaDe500(t *testing.T) {
	fx := setupChatFixture(t)
	sender := dialChatWS(t, fx.host, fx.aulaID, fx.alunoToken)
	peer := dialChatWS(t, fx.host, fx.aulaID, fx.profToken)

	writeChatJSON(t, sender, map[string]string{"type": "chat.send", "texto": "   "})
	empty := readChatFrame(t, sender)
	require.Equal(t, "chat.error", empty.Type)
	require.Equal(t, "mensagem vazia", empty.Error)

	tooLong := strings.Repeat("á", 501)
	writeChatJSON(t, sender, map[string]string{"type": "chat.send", "texto": tooLong})
	over := readChatFrame(t, sender)
	require.Equal(t, "chat.error", over.Type)
	require.Equal(t, "mensagem excede 500 caracteres", over.Error)

	ok500 := strings.Repeat("á", 500)
	writeChatJSON(t, sender, map[string]string{"type": "chat.send", "texto": ok500})
	msg := readChatFrame(t, sender)
	require.Equal(t, "chat.message", msg.Type)
	require.Equal(t, 500, len([]rune(msg.Texto)))
	peerMsg := readChatFrame(t, peer)
	require.Equal(t, "chat.message", peerMsg.Type)
	require.Equal(t, msg.ID, peerMsg.ID)
}

func TestChatTipoInvalido(t *testing.T) {
	fx := setupChatFixture(t)
	conn := dialChatWS(t, fx.host, fx.aulaID, fx.alunoToken)

	writeChatJSON(t, conn, map[string]string{"type": "chat.delete", "texto": "x"})
	frame := readChatFrame(t, conn)
	require.Equal(t, "chat.error", frame.Type)
	require.Equal(t, "tipo de mensagem inválido", frame.Error)
}

func TestChatIsolamentoDuasAulas(t *testing.T) {
	fx := setupChatFixture(t)
	salaA := dialChatWS(t, fx.host, fx.aulaID, fx.alunoToken)
	salaB := dialChatWS(t, fx.host, fx.aula2ID, fx.profToken)

	writeChatJSON(t, salaA, map[string]string{"type": "chat.send", "texto": "só na A"})
	gotA := readChatFrame(t, salaA)
	require.Equal(t, "chat.message", gotA.Type)
	require.Equal(t, fx.aulaID, gotA.AulaID)

	require.NoError(t, salaB.SetReadDeadline(time.Now().Add(150*time.Millisecond)))
	_, _, err := salaB.ReadMessage()
	require.Error(t, err)
}

func TestChatPingPong(t *testing.T) {
	fx := setupChatFixture(t)
	conn := dialChatWS(t, fx.host, fx.aulaID, fx.adminToken)

	writeChatJSON(t, conn, map[string]string{"type": "chat.ping"})
	frame := readChatFrame(t, conn)
	require.Equal(t, "chat.pong", frame.Type)
}

func TestChatIdentidadeVemDoServidor(t *testing.T) {
	fx := setupChatFixture(t)
	conn := dialChatWS(t, fx.host, fx.aulaID, fx.alunoToken)

	writeChatJSON(t, conn, map[string]any{
		"type":    "chat.send",
		"texto":   "oi",
		"autor":   map[string]any{"id": 999, "nome": "Falso"},
		"aula_id": fx.aula2ID,
	})
	frame := readChatFrame(t, conn)
	require.Equal(t, "chat.message", frame.Type)
	require.Equal(t, "Maria Silva", frame.Autor.Nome)
	require.Equal(t, fx.aulaID, frame.AulaID)
	require.NotEqual(t, uint(999), frame.Autor.ID)
}
