package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/auth"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/repository"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Handler struct {
	repos  *repository.Repositories
	tokens *auth.TokenService
}

func New(repos *repository.Repositories, cfg config.Config) *Handler {
	return &Handler{
		repos:  repos,
		tokens: auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration),
	}
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
