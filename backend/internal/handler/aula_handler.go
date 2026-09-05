package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"
	"github.com/gofiber/fiber/v2"
)

type aulaRequest struct {
	CursoID    uint   `json:"curso_id"`
	Titulo     string `json:"titulo"`
	Descricao  string `json:"descricao"`
	AgendadaEm string `json:"agendada_em"`
	Status     string `json:"status"`
}

func (h *Handler) ListAulas(c *fiber.Ctx) error {
	var cursoID *uint
	if q := c.Query("curso_id"); q != "" {
		id, err := strconv.ParseUint(q, 10, 64)
		if err != nil || id == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "curso_id inválido"})
		}
		v := uint(id)
		cursoID = &v
	}
	aulas, err := h.repos.Aulas.List(cursoID)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(aulas)
}

func (h *Handler) GetAula(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	aula, err := h.repos.Aulas.Get(id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(aula)
}

func (h *Handler) CreateAula(c *fiber.Ctx) error {
	var req aulaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if err := validateAulaRequest(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	agendadaEm, err := parseAgendadaEm(req.AgendadaEm)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	exists, err := h.repos.Aulas.CursoExists(req.CursoID)
	if err != nil {
		return handleError(c, err)
	}
	if !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "curso informado não existe"})
	}
	aula := models.Aula{
		CursoID:    req.CursoID,
		Titulo:     strings.TrimSpace(req.Titulo),
		Descricao:  req.Descricao,
		AgendadaEm: timeutil.NewDateTime(agendadaEm),
		Status:     normalizeAulaStatus(req.Status),
	}
	if err := h.repos.Aulas.Create(&aula); err != nil {
		return handleError(c, err)
	}
	created, err := h.repos.Aulas.Get(aula.ID)
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *Handler) UpdateAula(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var req aulaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}
	if err := validateAulaRequest(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	agendadaEm, err := parseAgendadaEm(req.AgendadaEm)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	exists, err := h.repos.Aulas.CursoExists(req.CursoID)
	if err != nil {
		return handleError(c, err)
	}
	if !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "curso informado não existe"})
	}
	aula := models.Aula{
		ID:         id,
		CursoID:    req.CursoID,
		Titulo:     strings.TrimSpace(req.Titulo),
		Descricao:  req.Descricao,
		AgendadaEm: timeutil.NewDateTime(agendadaEm),
		Status:     normalizeAulaStatus(req.Status),
	}
	if err := h.repos.Aulas.Update(&aula); err != nil {
		return handleError(c, err)
	}
	updated, err := h.repos.Aulas.Get(id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(updated)
}

func (h *Handler) DeleteAula(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	h.deleteVodForAula(id)
	h.deleteLiveForAula(id)
	if err := h.repos.Aulas.Delete(id); err != nil {
		return handleError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func validateAulaRequest(req aulaRequest) error {
	if req.CursoID == 0 {
		return errors.New("curso_id é obrigatório")
	}
	if strings.TrimSpace(req.Titulo) == "" {
		return errors.New("titulo é obrigatório")
	}
	if strings.TrimSpace(req.AgendadaEm) == "" {
		return errors.New("agendada_em é obrigatório")
	}
	return nil
}

func parseAgendadaEm(value string) (time.Time, error) {
	return timeutil.ParseDateTime(value)
}

func normalizeAulaStatus(status string) string {
	switch status {
	case models.AulaStatusAoVivo, models.AulaStatusEncerrada:
		return status
	default:
		return models.AulaStatusAgendada
	}
}
