package repository

import (
	"errors"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"gorm.io/gorm"
)

type VodRepository struct {
	db *gorm.DB
}

func (r *VodRepository) GetByAulaID(aulaID uint) (*models.AulaVod, error) {
	var vod models.AulaVod
	err := r.db.Where("aula_id = ?", aulaID).First(&vod).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &vod, err
}

func (r *VodRepository) Upsert(vod *models.AulaVod) error {
	existing, err := r.GetByAulaID(vod.AulaID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if errors.Is(err, ErrNotFound) {
		return r.db.Create(vod).Error
	}
	existing.StorageKey = vod.StorageKey
	existing.ContentType = vod.ContentType
	existing.SizeBytes = vod.SizeBytes
	existing.Status = vod.Status
	if err := r.db.Save(existing).Error; err != nil {
		return err
	}
	*vod = *existing
	return nil
}

func (r *VodRepository) DeleteByAulaID(aulaID uint) error {
	result := r.db.Where("aula_id = ?", aulaID).Delete(&models.AulaVod{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
