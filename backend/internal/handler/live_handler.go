package handler

import (
	"errors"
	"strings"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/repository"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetLive(c *fiber.Ctx) error {
	aulaID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if _, err := h.repos.Aulas.Get(aulaID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aula não encontrada"})
		}
		return handleError(c, err)
	}
	live, err := h.repos.Live.GetByAulaID(aulaID)
	if errors.Is(err, repository.ErrNotFound) {
		return c.JSON(liveStatusResponse(aulaID, h.liveBackend, nil))
	}
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(liveStatusResponse(aulaID, h.liveBackend, live))
}

func (h *Handler) StartLive(c *fiber.Ctx) error {
	aulaID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if _, err := h.repos.Aulas.Get(aulaID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aula não encontrada"})
		}
		return handleError(c, err)
	}
	if err := h.ensureLiveBackendReady(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error()})
	}

	live, err := h.repos.Live.Start(aulaID)
	if errors.Is(err, repository.ErrLiveAlreadyActive) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Já existe uma transmissão ao vivo. Encerre-a antes de iniciar.",
		})
	}
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(liveStatusResponse(aulaID, h.liveBackend, live))
}

func (h *Handler) ScheduleLive(c *fiber.Ctx) error {
	aulaID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	aula, err := h.repos.Aulas.Get(aulaID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aula não encontrada"})
		}
		return handleError(c, err)
	}
	if aula.AgendadaEm.Time.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Para agendar a transmissão, a aula precisa ter horário (agendada_em).",
		})
	}

	live, err := h.repos.Live.Schedule(aulaID)
	if errors.Is(err, repository.ErrLiveScheduleConflict) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Não é possível agendar enquanto a transmissão está ao vivo.",
		})
	}
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(liveStatusResponse(aulaID, h.liveBackend, live))
}

func (h *Handler) CancelLive(c *fiber.Ctx) error {
	aulaID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if _, err := h.repos.Aulas.Get(aulaID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aula não encontrada"})
		}
		return handleError(c, err)
	}

	err = h.repos.Live.Cancel(aulaID)
	if errors.Is(err, repository.ErrLiveNotScheduled) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Esta aula não tem transmissão agendada para cancelar.",
		})
	}
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(liveStatusResponse(aulaID, h.liveBackend, nil))
}

func (h *Handler) StopLive(c *fiber.Ctx) error {
	aulaID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if _, err := h.repos.Aulas.Get(aulaID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aula não encontrada"})
		}
		return handleError(c, err)
	}

	live, err := h.repos.Live.Stop(aulaID)
	if errors.Is(err, repository.ErrLiveNotActive) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Esta aula não está com transmissão ao vivo.",
		})
	}
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(liveStatusResponse(aulaID, h.liveBackend, live))
}

func (h *Handler) GetLivePlayback(c *fiber.Ctx) error {
	aulaID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if _, err := h.repos.Aulas.Get(aulaID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aula não encontrada"})
		}
		return handleError(c, err)
	}
	live, err := h.repos.Live.GetByAulaID(aulaID)
	if err != nil || live.Status != models.LiveStatusAoVivo {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Não há transmissão ao vivo nesta aula.",
		})
	}

	modo := h.liveBackend
	if strings.EqualFold(modo, "ivs") {
		if !h.ivsReady() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Transmissão IVS não configurada neste ambiente.",
			})
		}
		return c.JSON(fiber.Map{
			"aula_id":      aulaID,
			"modo":         "ivs",
			"protocolo":    "hls",
			"player":       "ivs",
			"playback_url": h.ivsPlaybackURL,
		})
	}
	return c.JSON(fiber.Map{
		"aula_id":      aulaID,
		"modo":         "stub",
		"protocolo":    nil,
		"player":       nil,
		"playback_url": nil,
		"mensagem":     "Ambiente local: não há sinal de vídeo real.",
	})
}

func (h *Handler) GetLiveIngest(c *fiber.Ctx) error {
	aulaID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if _, err := h.repos.Aulas.Get(aulaID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aula não encontrada"})
		}
		return handleError(c, err)
	}
	live, err := h.repos.Live.GetByAulaID(aulaID)
	if err != nil || live.Status != models.LiveStatusAoVivo {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Não há transmissão ao vivo nesta aula.",
		})
	}

	if strings.EqualFold(h.liveBackend, "ivs") {
		if !h.ivsReady() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Transmissão IVS não configurada neste ambiente.",
			})
		}
		return c.JSON(fiber.Map{
			"aula_id":       aulaID,
			"modo":          "ivs",
			"ingest_server": h.ivsIngestServer(),
			"stream_key":    h.ivsStreamKey,
			"observacao":    "A mesma stream key vale até o destroy da sessão Terraform.",
		})
	}
	return c.JSON(fiber.Map{
		"aula_id":       aulaID,
		"modo":          "stub",
		"ingest_server": nil,
		"stream_key":    nil,
		"mensagem":      "Ingestão real só na sessão AWS.",
	})
}

func (h *Handler) ensureLiveBackendReady() error {
	if !strings.EqualFold(h.liveBackend, "ivs") {
		return nil
	}
	if h.ivsReady() {
		return nil
	}
	return errors.New("Transmissão IVS não configurada neste ambiente.")
}

func (h *Handler) ivsReady() bool {
	return strings.TrimSpace(h.ivsIngestEndpoint) != "" &&
		strings.TrimSpace(h.ivsStreamKey) != "" &&
		strings.TrimSpace(h.ivsPlaybackURL) != ""
}

func (h *Handler) ivsIngestServer() string {
	host := strings.TrimSpace(h.ivsIngestEndpoint)
	host = strings.TrimPrefix(host, "rtmps://")
	host = strings.TrimPrefix(host, "rtmp://")
	host = strings.TrimSuffix(host, "/")
	if i := strings.Index(host, "/"); i >= 0 {
		host = host[:i]
	}
	return "rtmps://" + host + ":443/app/"
}

func liveStatusResponse(aulaID uint, modo string, live *models.AulaLive) fiber.Map {
	modo = strings.ToLower(strings.TrimSpace(modo))
	if modo == "" {
		modo = "stub"
	}
	resp := fiber.Map{
		"aula_id":      aulaID,
		"status":       models.LiveStatusInativa,
		"modo":         modo,
		"iniciada_em":  nil,
		"encerrada_em": nil,
	}
	if live == nil {
		return resp
	}
	resp["status"] = live.Status
	if live.IniciadaEm != nil && !live.IniciadaEm.Time.IsZero() {
		resp["iniciada_em"] = timeutil.FormatDateTime(live.IniciadaEm.Time)
	}
	if live.EncerradaEm != nil && !live.EncerradaEm.Time.IsZero() {
		resp["encerrada_em"] = timeutil.FormatDateTime(live.EncerradaEm.Time)
	}
	return resp
}

func (h *Handler) deleteLiveForAula(aulaID uint) {
	_ = h.repos.Live.DeleteByAulaID(aulaID)
}
