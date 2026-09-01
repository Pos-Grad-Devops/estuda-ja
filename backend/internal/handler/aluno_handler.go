package handler

import (
	"errors"
	"strings"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

type alunoRequest struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
}

func (h *Handler) ListAlunos(c *fiber.Ctx) error {
	alunos, err := h.repos.Alunos.List()
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(alunos)
}

func (h *Handler) GetAluno(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	aluno, err := h.repos.Alunos.Get(id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(aluno)
}

func (h *Handler) CreateAluno(c *fiber.Ctx) error {
	var req alunoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if err := validateAlunoRequest(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	aluno := models.Aluno{Nome: strings.TrimSpace(req.Nome), Email: strings.TrimSpace(req.Email)}
	if err := h.repos.Alunos.Create(&aluno); err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(aluno)
}

func (h *Handler) UpdateAluno(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var req alunoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if err := validateAlunoRequest(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	aluno := models.Aluno{ID: id, Nome: strings.TrimSpace(req.Nome), Email: strings.TrimSpace(req.Email)}
	if err := h.repos.Alunos.Update(&aluno); err != nil {
		return handleError(c, err)
	}
	updated, err := h.repos.Alunos.Get(id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(updated)
}

func (h *Handler) DeleteAluno(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.repos.Alunos.Delete(id); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func validateAlunoRequest(req alunoRequest) error {
	if strings.TrimSpace(req.Nome) == "" {
		return errors.New("nome é obrigatório")
	}
	if strings.TrimSpace(req.Email) == "" {
		return errors.New("email é obrigatório")
	}
	return nil
}
