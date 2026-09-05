package handler

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/auth"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/chat"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/repository"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/vodstorage"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Handler struct {
	repos             *repository.Repositories
	tokens            *auth.TokenService
	vod               vodstorage.Storage
	vodPlaybackTTL    time.Duration
	jwtSecret         string
	liveBackend       string
	ivsIngestEndpoint string
	ivsStreamKey      string
	ivsPlaybackURL    string
	ivsChannelARN     string
	chatHub           *chat.Hub
}

func New(repos *repository.Repositories, cfg config.Config) *Handler {
	h := &Handler{
		repos:             repos,
		tokens:            auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration),
		vodPlaybackTTL:    cfg.VODPlaybackTTL,
		jwtSecret:         cfg.JWTSecret,
		liveBackend:       strings.ToLower(strings.TrimSpace(cfg.LiveBackend)),
		ivsIngestEndpoint: cfg.IVSIngestEndpoint,
		ivsStreamKey:      cfg.IVSStreamKey,
		ivsPlaybackURL:    cfg.IVSPlaybackURL,
		ivsChannelARN:     cfg.IVSChannelARN,
		chatHub:           chat.NewHub(),
	}
	if h.liveBackend == "" {
		h.liveBackend = "stub"
	}
	if h.vodPlaybackTTL <= 0 {
		h.vodPlaybackTTL = 15 * time.Minute
	}
	storage, err := vodstorage.NewFromConfig(context.Background(), cfg)
	if err != nil {
		if strings.EqualFold(strings.TrimSpace(cfg.VODBackend), "s3") {
			log.Fatalf("VOD storage: %v", err)
		}
		log.Printf("aviso VOD storage: %v", err)
	} else {
		h.vod = storage
	}
	return h
}

// WithStorage substitui o backend de storage (útil em testes).
func (h *Handler) WithStorage(s vodstorage.Storage) *Handler {
	h.vod = s
	return h
}

func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func parseID(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "id inválido")
	}
	return uint(id), nil
}

func handleError(c *fiber.Ctx, err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "não encontrado"})
	}
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "curso informado não existe"})
	}
	if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "registro duplicado"})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro interno"})
}
