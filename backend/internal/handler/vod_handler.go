package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/repository"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/vodstorage"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type vodContentClaims struct {
	AulaID  uint   `json:"aula_id"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

const vodContentPurpose = "vod_content"

func (h *Handler) GetVod(c *fiber.Ctx) error {
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
	vod, err := h.repos.Vod.GetByAulaID(aulaID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Gravação não encontrada"})
		}
		return handleError(c, err)
	}
	if vod.Status != models.VodStatusPublicado {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Gravação não encontrada"})
	}
	return c.JSON(vodPublicResponse(vod))
}

func (h *Handler) GetVodPlayback(c *fiber.Ctx) error {
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
	vod, err := h.repos.Vod.GetByAulaID(aulaID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Gravação não encontrada"})
		}
		return handleError(c, err)
	}
	if vod.Status != models.VodStatusPublicado {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Gravação não encontrada"})
	}

	ttl := h.vodPlaybackTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	var playbackURL string
	var expiresAt time.Time

	if presigner, ok := h.vod.(vodstorage.Presigner); ok {
		url, exp, err := presigner.PresignGet(c.Context(), vod.StorageKey, ttl)
		if err != nil {
			log.Printf("vod presign: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao gerar URL de reprodução"})
		}
		playbackURL = url
		expiresAt = exp
	} else {
		token, err := h.signVodContentToken(aulaID, ttl)
		if err != nil {
			return handleError(c, err)
		}
		expiresAt = time.Now().Add(ttl)
		playbackURL = fmt.Sprintf("%s/api/v1/aulas/%d/vod/content?token=%s", c.BaseURL(), aulaID, token)
	}

	return c.JSON(fiber.Map{
		"playback_url":       playbackURL,
		"expires_at":         timeutil.FormatDateTime(expiresAt),
		"expires_in_seconds": int(ttl.Seconds()),
	})
}

func (h *Handler) GetVodContent(c *fiber.Ctx) error {
	aulaID, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	tokenStr := c.Query("token")
	if tokenStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token de reprodução obrigatório"})
	}
	claims, err := h.parseVodContentToken(tokenStr)
	if err != nil || claims.AulaID != aulaID || claims.Purpose != vodContentPurpose {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token de reprodução inválido ou expirado"})
	}

	vod, err := h.repos.Vod.GetByAulaID(aulaID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Gravação não encontrada"})
		}
		return handleError(c, err)
	}
	if h.vod == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "armazenamento VOD não configurado"})
	}

	rc, size, err := h.vod.Open(c.Context(), vod.StorageKey)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Gravação não encontrada"})
	}
	defer rc.Close()

	c.Set("Content-Type", models.VodContentTypeMP4)
	c.Set("Content-Disposition", "inline")
	c.Set("Content-Length", strconv.FormatInt(size, 10))
	c.Status(fiber.StatusOK)
	_, err = io.Copy(c, rc)
	return err
}

func (h *Handler) PutVod(c *fiber.Ctx) error {
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
	if h.vod == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "armazenamento VOD não configurado"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "arquivo é obrigatório"})
	}
	if err := vodstorage.ValidateUpload(fileHeader, fileHeader.Open); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "não foi possível ler o arquivo"})
	}
	defer src.Close()

	key := vodstorage.ObjectKey(aulaID)
	ctx := context.Background()
	if err := h.vod.Put(ctx, key, src, fileHeader.Size, models.VodContentTypeMP4); err != nil {
		log.Printf("vod put: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao salvar gravação"})
	}

	vod := &models.AulaVod{
		AulaID:      aulaID,
		StorageKey:  key,
		ContentType: models.VodContentTypeMP4,
		SizeBytes:   fileHeader.Size,
		Status:      models.VodStatusPublicado,
	}
	if err := h.repos.Vod.Upsert(vod); err != nil {
		_ = h.vod.Delete(ctx, key)
		return handleError(c, err)
	}

	saved, err := h.repos.Vod.GetByAulaID(aulaID)
	if err != nil {
		return handleError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(vodPublicResponse(saved))
}

func (h *Handler) DeleteVod(c *fiber.Ctx) error {
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
	vod, err := h.repos.Vod.GetByAulaID(aulaID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Gravação não encontrada"})
		}
		return handleError(c, err)
	}
	if err := h.repos.Vod.DeleteByAulaID(aulaID); err != nil {
		return handleError(c, err)
	}
	if h.vod != nil {
		if err := h.vod.Delete(context.Background(), vod.StorageKey); err != nil {
			log.Printf("vod delete storage: %v", err)
		}
	}
	return c.JSON(fiber.Map{"message": "gravação removida"})
}

func (h *Handler) deleteVodForAula(aulaID uint) {
	vod, err := h.repos.Vod.GetByAulaID(aulaID)
	if err != nil {
		return
	}
	_ = h.repos.Vod.DeleteByAulaID(aulaID)
	if h.vod != nil {
		if err := h.vod.Delete(context.Background(), vod.StorageKey); err != nil {
			log.Printf("vod cascade delete: %v", err)
		}
	}
}

func vodPublicResponse(vod *models.AulaVod) fiber.Map {
	return fiber.Map{
		"aula_id":      vod.AulaID,
		"status":       vod.Status,
		"content_type": vod.ContentType,
		"size_bytes":   vod.SizeBytes,
		"updated_at":   timeutil.FormatDateTime(vod.UpdatedAt.Time),
	}
}

func (h *Handler) signVodContentToken(aulaID uint, ttl time.Duration) (string, error) {
	claims := vodContentClaims{
		AulaID:  aulaID,
		Purpose: vodContentPurpose,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

func (h *Handler) parseVodContentToken(tokenString string) (*vodContentClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &vodContentClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*vodContentClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}
	return claims, nil
}
