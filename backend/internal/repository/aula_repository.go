package repository

import (
	"errors"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"gorm.io/gorm"
)

type AulaRepository struct {
	db *gorm.DB
}

func (r *AulaRepository) List(cursoID *uint) ([]models.Aula, error) {
	var aulas []models.Aula
	q := r.db.Preload("Curso").Order("agendada_em asc")
	if cursoID != nil {
		q = q.Where("curso_id = ?", *cursoID)
	}
	err := q.Find(&aulas).Error
	return aulas, err
}

func (r *AulaRepository) Get(id uint) (*models.Aula, error) {
	var aula models.Aula
	err := r.db.Preload("Curso").First(&aula, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &aula, err
}

func (r *AulaRepository) Create(aula *models.Aula) error {
	return r.db.Create(aula).Error
}

func (r *AulaRepository) Update(aula *models.Aula) error {
	result := r.db.Model(&models.Aula{}).Where("id = ?", aula.ID).Updates(map[string]interface{}{
		"curso_id":    aula.CursoID,
		"titulo":      aula.Titulo,
		"descricao":   aula.Descricao,
		"agendada_em": aula.AgendadaEm.Time,
		"status":      aula.Status,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AulaRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Aula{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AulaRepository) CursoExists(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Curso{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
