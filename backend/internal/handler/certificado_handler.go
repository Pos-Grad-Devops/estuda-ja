package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/certpdf"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/middleware"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/repository"
	"github.com/gofiber/fiber/v2"
)

type elegibilidadeRequest struct {
	UserID uint `json:"user_id"`
}

type certificadoResumo struct {
	ID          uint   `json:"id"`
	Status      string `json:"status"`
	EmitidoEm   any    `json:"emitido_em"`
	AlunoNome   string `json:"aluno_nome"`
	CursoTitulo string `json:"curso_titulo"`
}

func certificadoToResumo(c *models.Certificado) certificadoResumo {
	return certificadoResumo{
		ID:          c.ID,
		Status:      c.Status,
		EmitidoEm:   c.EmitidoEm,
		AlunoNome:   c.AlunoNome,
		CursoTitulo: c.CursoTitulo,
	}
}

func parseCertID(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("certId"), 10, 64)
	if err != nil || id == 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "id inválido")
	}
	return uint(id), nil
}

func (h *Handler) ensureCursoExists(cursoID uint) error {
	_, err := h.repos.Cursos.Get(cursoID)
	return err
}

// GetCertificadoStatus — GET /cursos/:id/certificado
// Aluno: próprio. Admin: próprio ou ?user_id=.
func (h *Handler) GetCertificadoStatus(c *fiber.Ctx) error {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}

	cursoID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.ensureCursoExists(cursoID); err != nil {
		return handleError(c, err)
	}

	targetUserID := claims.UserID
	if q := strings.TrimSpace(c.Query("user_id")); q != "" {
		if claims.Role != models.RoleAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "sem permissão"})
		}
		uid, err := strconv.ParseUint(q, 10, 64)
		if err != nil || uid == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id inválido"})
		}
		targetUserID = uint(uid)
	} else if claims.Role == models.RoleProfessor {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "sem permissão"})
	} else if claims.Role != models.RoleAluno && claims.Role != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "sem permissão"})
	}

	elegivel, err := h.repos.Certificados.IsElegivel(targetUserID, cursoID)
	if err != nil {
		return handleError(c, err)
	}

	var certPayload any
	if valido, err := h.repos.Certificados.FindValido(targetUserID, cursoID); err == nil {
		resumo := certificadoToResumo(valido)
		certPayload = resumo
	} else if errors.Is(err, repository.ErrNotFound) {
		if recente, err2 := h.repos.Certificados.FindMaisRecente(targetUserID, cursoID); err2 == nil {
			resumo := certificadoToResumo(recente)
			certPayload = resumo
		} else if !errors.Is(err2, repository.ErrNotFound) {
			return handleError(c, err2)
		}
	} else {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"curso_id":    cursoID,
		"user_id":     targetUserID,
		"elegivel":    elegivel,
		"certificado": certPayload,
	})
}

// PutCertificadoElegibilidade — PUT /cursos/:id/certificados/elegibilidade (só admin).
// Marca/reabilita elegibilidade; NÃO emite certificado.
func (h *Handler) PutCertificadoElegibilidade(c *fiber.Ctx) error {
	cursoID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.ensureCursoExists(cursoID); err != nil {
		return handleError(c, err)
	}

	var req elegibilidadeRequest
	if err := c.BodyParser(&req); err != nil || req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}

	user, err := h.repos.Users.Get(req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "usuário não encontrado"})
		}
		return handleError(c, err)
	}
	if user.Role != models.RoleAluno {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "usuário não é aluno"})
	}

	if _, err := h.repos.Certificados.UpsertElegibilidade(req.UserID, cursoID); err != nil {
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"curso_id": cursoID,
		"user_id":  req.UserID,
		"elegivel": true,
	})
}

// GetCertificadoPDF — GET /cursos/:id/certificado/pdf (só o próprio aluno).
// Lazy emit se elegível sem ativo; reutiliza valido; 403 se não elegível.
func (h *Handler) GetCertificadoPDF(c *fiber.Ctx) error {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
	}
	if claims.Role != models.RoleAluno {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "sem permissão"})
	}

	cursoID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.ensureCursoExists(cursoID); err != nil {
		return handleError(c, err)
	}

	userID := claims.UserID
	valido, err := h.repos.Certificados.FindValido(userID, cursoID)
	if err == nil {
		return h.streamCertificadoPDF(c, cursoID, valido)
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return handleError(c, err)
	}

	elegivel, err := h.repos.Certificados.IsElegivel(userID, cursoID)
	if err != nil {
		return handleError(c, err)
	}
	if !elegivel {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Você não está elegível para o certificado deste curso",
		})
	}

	user, err := h.repos.Users.Get(userID)
	if err != nil {
		return handleError(c, err)
	}
	curso, err := h.repos.Cursos.Get(cursoID)
	if err != nil {
		return handleError(c, err)
	}

	created, err := h.repos.Certificados.CreateValido(userID, cursoID, user.Nome, curso.Titulo)
	if err != nil {
		return handleError(c, err)
	}

	pdfBytes, err := certpdf.Generate(created.AlunoNome, created.CursoTitulo, created.EmitidoEm)
	if err != nil {
		_ = h.repos.Certificados.DeleteByID(created.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Não foi possível gerar o certificado",
		})
	}

	filename := fmt.Sprintf("certificado-curso-%d.pdf", cursoID)
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.Send(pdfBytes)
}

func (h *Handler) streamCertificadoPDF(c *fiber.Ctx, cursoID uint, cert *models.Certificado) error {
	pdfBytes, err := certpdf.Generate(cert.AlunoNome, cert.CursoTitulo, cert.EmitidoEm)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Não foi possível gerar o certificado",
		})
	}
	filename := fmt.Sprintf("certificado-curso-%d.pdf", cursoID)
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.Send(pdfBytes)
}

// ListCertificadosCurso — GET /cursos/:id/certificados (só admin).
func (h *Handler) ListCertificadosCurso(c *fiber.Ctx) error {
	cursoID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.ensureCursoExists(cursoID); err != nil {
		return handleError(c, err)
	}

	elegs, err := h.repos.Certificados.ListElegibilidadesByCurso(cursoID)
	if err != nil {
		return handleError(c, err)
	}
	certs, err := h.repos.Certificados.ListByCurso(cursoID)
	if err != nil {
		return handleError(c, err)
	}

	elegPayload := make([]fiber.Map, 0, len(elegs))
	for _, e := range elegs {
		nome := ""
		if u, err := h.repos.Users.Get(e.UserID); err == nil {
			nome = u.Nome
		}
		elegPayload = append(elegPayload, fiber.Map{
			"user_id":    e.UserID,
			"user_nome":  nome,
			"created_at": e.CreatedAt,
		})
	}

	certPayload := make([]fiber.Map, 0, len(certs))
	for _, cert := range certs {
		certPayload = append(certPayload, fiber.Map{
			"id":           cert.ID,
			"user_id":      cert.UserID,
			"status":       cert.Status,
			"emitido_em":   cert.EmitidoEm,
			"aluno_nome":   cert.AlunoNome,
			"curso_titulo": cert.CursoTitulo,
		})
	}

	return c.JSON(fiber.Map{
		"curso_id":       cursoID,
		"elegibilidades": elegPayload,
		"certificados":   certPayload,
	})
}

// InvalidarCertificado — POST /cursos/:id/certificados/:certId/invalidar (só admin).
func (h *Handler) InvalidarCertificado(c *fiber.Ctx) error {
	cursoID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	certID, err := parseCertID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.ensureCursoExists(cursoID); err != nil {
		return handleError(c, err)
	}

	cert, err := h.repos.Certificados.Invalidar(cursoID, certID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "certificado não encontrado"})
		}
		if errors.Is(err, repository.ErrCertificadoNaoPertenceCurso) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "certificado não pertence a este curso"})
		}
		if errors.Is(err, repository.ErrCertificadoJaInvalidado) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "certificado já invalidado"})
		}
		return handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"id":       cert.ID,
		"status":   cert.Status,
		"curso_id": cert.CursoID,
		"user_id":  cert.UserID,
	})
}
