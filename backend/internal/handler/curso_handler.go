package handler

import (
	"strings"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

type cursoRequest struct {
	Titulo    string `json:"titulo"`
	Descricao string `json:"descricao"`
}

func (h *Handler) ListCursos(c *fiber.Ctx) error {
	cursos, err := h.repos.Cursos.List()
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(cursos)
}

func (h *Handler) GetCurso(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	curso, err := h.repos.Cursos.Get(id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(curso)
}

func (h *Handler) CreateCurso(c *fiber.Ctx) error {
	var req cursoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if strings.TrimSpace(req.Titulo) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "titulo é obrigatório"})
	}
	curso := models.Curso{Titulo: strings.TrimSpace(req.Titulo), Descricao: req.Descricao}
	if err := h.repos.Cursos.Create(&curso); err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(curso)
}

func (h *Handler) UpdateCurso(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var req cursoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if strings.TrimSpace(req.Titulo) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "titulo é obrigatório"})
	}
	curso := models.Curso{ID: id, Titulo: strings.TrimSpace(req.Titulo), Descricao: req.Descricao}
	if err := h.repos.Cursos.Update(&curso); err != nil {
		return handleError(c, err)
	}
	updated, err := h.repos.Cursos.Get(id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(updated)
}

func (h *Handler) DeleteCurso(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.repos.Cursos.Delete(id); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
