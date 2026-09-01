package handler

import (
	"errors"
	"strings"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/auth"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

type userRequest struct {
	Nome     string      `json:"nome"`
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Role     models.Role `json:"role"`
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	users, err := h.repos.Users.List()
	if err != nil {
		return handleError(c, err)
	}

	public := make([]models.User, 0, len(users))
	for _, user := range users {
		public = append(public, user.Public())
	}
	return c.JSON(public)
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.repos.Users.Get(id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(user.Public())
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req userRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if err := validateUserRequest(req, true); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return handleError(c, err)
	}

	user := models.User{
		Nome:         strings.TrimSpace(req.Nome),
		Email:        strings.TrimSpace(req.Email),
		PasswordHash: hash,
		Role:         req.Role,
	}
	if err := h.repos.Users.Create(&user); err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(user.Public())
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var req userRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if err := validateUserRequest(req, false); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user := models.User{
		ID:    id,
		Nome:  strings.TrimSpace(req.Nome),
		Email: strings.TrimSpace(req.Email),
		Role:  req.Role,
	}
	if err := h.repos.Users.Update(&user); err != nil {
		return handleError(c, err)
	}

	if req.Password != "" {
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			return handleError(c, err)
		}
		if err := h.repos.Users.UpdatePassword(id, hash); err != nil {
			return handleError(c, err)
		}
	}

	updated, err := h.repos.Users.Get(id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(updated.Public())
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.repos.Users.Delete(id); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func validateUserRequest(req userRequest, requirePassword bool) error {
	if strings.TrimSpace(req.Nome) == "" {
		return errors.New("nome é obrigatório")
	}
	if strings.TrimSpace(req.Email) == "" {
		return errors.New("email é obrigatório")
	}
	if requirePassword && req.Password == "" {
		return errors.New("senha é obrigatória")
	}
	switch req.Role {
	case models.RoleAdmin, models.RoleProfessor, models.RoleAluno:
		return nil
	default:
		return errors.New("perfil inválido")
	}
}
