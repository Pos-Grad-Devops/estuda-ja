package handler

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/chat"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/repository"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	chatMaxRunes     = 500
	chatPingInterval = 2 * time.Minute

	chatLocalAulaID = "chatAulaID"
	chatLocalUserID = "chatUserID"
	chatLocalNome   = "chatNome"
	chatLocalRole   = "chatRole"
)

type chatIncoming struct {
	Type  string `json:"type"`
	Texto string `json:"texto"`
}

type chatAutor struct {
	ID   uint   `json:"id"`
	Nome string `json:"nome"`
}

type chatMessageFrame struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	AulaID    uint      `json:"aula_id"`
	Autor     chatAutor `json:"autor"`
	Texto     string    `json:"texto"`
	EnviadoEm string    `json:"enviado_em"`
}

type chatErrorFrame struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

type chatPingPongFrame struct {
	Type string `json:"type"`
}

func canReadAulas(role models.Role) bool {
	switch role {
	case models.RoleAdmin, models.RoleProfessor, models.RoleAluno:
		return true
	default:
		return false
	}
}

// ChatHandshake autentica ?token= (JWT), valida a aula e carrega o nome.
// Não registra a query string (token não vai para o log).
func (h *Handler) ChatHandshake(c *fiber.Ctx) error {
	aulaID, err := parseID(c)
	if err != nil {
		log.Printf("chat handshake recusado motivo=id inválido")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id inválido"})
	}

	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		log.Printf("chat handshake recusado aula_id=%d motivo=token ausente", aulaID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}

	claims, err := h.tokens.Parse(token)
	if err != nil {
		log.Printf("chat handshake recusado aula_id=%d motivo=token inválido", aulaID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token inválido"})
	}

	if !canReadAulas(claims.Role) {
		log.Printf("chat handshake recusado aula_id=%d motivo=sem permissão", aulaID)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "sem permissão"})
	}

	if _, err := h.repos.Aulas.Get(aulaID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			log.Printf("chat handshake recusado aula_id=%d motivo=aula não encontrada", aulaID)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aula não encontrada"})
		}
		log.Printf("chat handshake recusado aula_id=%d motivo=erro interno", aulaID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro interno"})
	}

	user, err := h.repos.Users.Get(claims.UserID)
	if err != nil {
		log.Printf("chat handshake recusado aula_id=%d motivo=usuário não encontrado", aulaID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}

	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	c.Locals(chatLocalAulaID, aulaID)
	c.Locals(chatLocalUserID, claims.UserID)
	c.Locals(chatLocalNome, user.Nome)
	c.Locals(chatLocalRole, claims.Role)
	c.Locals("allowed", true)
	return c.Next()
}

func (h *Handler) ChatWS() fiber.Handler {
	return websocket.New(h.HandleChatConn)
}

func (h *Handler) HandleChatConn(conn *websocket.Conn) {
	aulaID, _ := conn.Locals(chatLocalAulaID).(uint)
	userID, _ := conn.Locals(chatLocalUserID).(uint)
	nome, _ := conn.Locals(chatLocalNome).(string)

	client := chat.NewClient(aulaID, userID, nome)
	h.chatHub.Register(client)
	defer func() {
		h.chatHub.Unregister(client)
		_ = conn.Close()
	}()

	go h.chatWritePump(client, conn)
	h.chatReadLoop(client, conn)
}

func (h *Handler) chatWritePump(client *chat.Client, conn *websocket.Conn) {
	ticker := time.NewTicker(chatPingInterval)
	defer ticker.Stop()

	for {
		select {
		case payload, ok := <-client.Send:
			if !ok {
				return
			}
			if err := writeChatText(client, conn, payload); err != nil {
				_ = conn.Close()
				return
			}
		case <-ticker.C:
			ping, _ := json.Marshal(chatPingPongFrame{Type: "chat.ping"})
			if err := writeChatText(client, conn, ping); err != nil {
				_ = conn.Close()
				return
			}
		}
	}
}

func (h *Handler) chatReadLoop(client *chat.Client, conn *websocket.Conn) {
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var in chatIncoming
		if err := json.Unmarshal(raw, &in); err != nil {
			h.sendChatError(client, "tipo de mensagem inválido")
			continue
		}

		switch in.Type {
		case "chat.send":
			h.handleChatSend(client, in.Texto)
		case "chat.ping":
			pong, _ := json.Marshal(chatPingPongFrame{Type: "chat.pong"})
			select {
			case client.Send <- pong:
			default:
			}
		case "chat.pong":
			// keepalive do cliente — sem ação
		default:
			h.sendChatError(client, "tipo de mensagem inválido")
		}
	}
}

func (h *Handler) handleChatSend(client *chat.Client, texto string) {
	texto = strings.TrimSpace(texto)
	n := utf8.RuneCountInString(texto)
	if n == 0 {
		h.sendChatError(client, "mensagem vazia")
		return
	}
	if n > chatMaxRunes {
		h.sendChatError(client, "mensagem excede 500 caracteres")
		return
	}

	frame := chatMessageFrame{
		Type:   "chat.message",
		ID:     uuid.NewString(),
		AulaID: client.AulaID,
		Autor: chatAutor{
			ID:   client.UserID,
			Nome: client.Nome,
		},
		Texto:     texto,
		EnviadoEm: time.Now().Format(timeutil.DateTimeLayout),
	}
	payload, err := json.Marshal(frame)
	if err != nil {
		h.sendChatError(client, "falha ao enviar mensagem")
		return
	}
	h.chatHub.Broadcast(client.AulaID, payload)
}

func (h *Handler) sendChatError(client *chat.Client, msg string) {
	payload, err := json.Marshal(chatErrorFrame{Type: "chat.error", Error: msg})
	if err != nil {
		return
	}
	select {
	case client.Send <- payload:
	default:
	}
}

func writeChatText(client *chat.Client, conn *websocket.Conn, payload []byte) error {
	return client.Write(func() error {
		return conn.WriteMessage(websocket.TextMessage, payload)
	})
}
