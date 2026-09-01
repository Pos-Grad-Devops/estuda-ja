package repository

import (
	"errors"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"gorm.io/gorm"
)

type CursoRepository struct {
	db *gorm.DB
}

func (r *CursoRepository) List() ([]models.Curso, error) {
	var cursos []models.Curso
	err := r.db.Order("id asc").Find(&cursos).Error
	return cursos, err
}

func (r *CursoRepository) Get(id uint) (*models.Curso, error) {
	var curso models.Curso
	err := r.db.Preload("Aulas").First(&curso, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &curso, err
}

func (r *CursoRepository) Create(curso *models.Curso) error {
	return r.db.Create(curso).Error
}

func (r *CursoRepository) Update(curso *models.Curso) error {
	result := r.db.Model(&models.Curso{}).Where("id = ?", curso.ID).Updates(map[string]interface{}{
		"titulo":    curso.Titulo,
		"descricao": curso.Descricao,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CursoRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Curso{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
