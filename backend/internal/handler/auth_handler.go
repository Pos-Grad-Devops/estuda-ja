package handler

import (
	"strings"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/auth"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}

	email := strings.TrimSpace(req.Email)
	if email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email e senha são obrigatórios"})
	}

	user, err := h.repos.Users.GetByEmail(email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "credenciais inválidas"})
	}
	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "credenciais inválidas"})
	}

	token, err := h.tokens.Generate(user)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user":  user.Public(),
	})
}

func (h *Handler) Me(c *fiber.Ctx) error {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}

	user, err := h.repos.Users.Get(claims.UserID)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(user.Public())
}
