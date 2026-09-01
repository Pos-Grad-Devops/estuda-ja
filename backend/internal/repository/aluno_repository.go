package repository

import (
	"errors"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"gorm.io/gorm"
)

type AlunoRepository struct {
	db *gorm.DB
}

func (r *AlunoRepository) List() ([]models.Aluno, error) {
	var alunos []models.Aluno
	err := r.db.Order("nome asc").Find(&alunos).Error
	return alunos, err
}

func (r *AlunoRepository) Get(id uint) (*models.Aluno, error) {
	var aluno models.Aluno
	err := r.db.First(&aluno, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &aluno, err
}

func (r *AlunoRepository) Create(aluno *models.Aluno) error {
	return r.db.Create(aluno).Error
}

func (r *AlunoRepository) Update(aluno *models.Aluno) error {
	result := r.db.Model(&models.Aluno{}).Where("id = ?", aluno.ID).Updates(map[string]interface{}{
		"nome":  aluno.Nome,
		"email": aluno.Email,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AlunoRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Aluno{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
