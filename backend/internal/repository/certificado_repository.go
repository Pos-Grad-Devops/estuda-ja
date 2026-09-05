package repository

import (
	"errors"
	"time"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"
	"gorm.io/gorm"
)

var (
	// ErrCertificadoJaInvalidado — certificado já não está válido.
	ErrCertificadoJaInvalidado = errors.New("certificado já invalidado")
	// ErrCertificadoNaoPertenceCurso — certId não pertence ao curso da rota.
	ErrCertificadoNaoPertenceCurso = errors.New("certificado não pertence a este curso")
)

type CertificadoRepository struct {
	db *gorm.DB
}

func (r *CertificadoRepository) GetElegibilidade(userID, cursoID uint) (*models.CertificadoElegibilidade, error) {
	var e models.CertificadoElegibilidade
	err := r.db.Where("user_id = ? AND curso_id = ?", userID, cursoID).First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &e, err
}

func (r *CertificadoRepository) IsElegivel(userID, cursoID uint) (bool, error) {
	_, err := r.GetElegibilidade(userID, cursoID)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpsertElegibilidade marca ou reabilita elegibilidade (idempotente). Não cria Certificado.
func (r *CertificadoRepository) UpsertElegibilidade(userID, cursoID uint) (*models.CertificadoElegibilidade, error) {
	existing, err := r.GetElegibilidade(userID, cursoID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	e := models.CertificadoElegibilidade{UserID: userID, CursoID: cursoID}
	if err := r.db.Create(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *CertificadoRepository) DeleteElegibilidade(userID, cursoID uint) error {
	result := r.db.Where("user_id = ? AND curso_id = ?", userID, cursoID).Delete(&models.CertificadoElegibilidade{})
	return result.Error
}

// FindValido retorna o certificado ativo (status=valido) do par, se houver.
func (r *CertificadoRepository) FindValido(userID, cursoID uint) (*models.Certificado, error) {
	var c models.Certificado
	err := r.db.Where("user_id = ? AND curso_id = ? AND status = ?", userID, cursoID, models.CertificadoStatusValido).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

// FindMaisRecente retorna o certificado mais recente do par (valido ou invalidado).
func (r *CertificadoRepository) FindMaisRecente(userID, cursoID uint) (*models.Certificado, error) {
	var c models.Certificado
	err := r.db.Where("user_id = ? AND curso_id = ?", userID, cursoID).
		Order("id desc").First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *CertificadoRepository) GetByID(id uint) (*models.Certificado, error) {
	var c models.Certificado
	err := r.db.First(&c, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *CertificadoRepository) ListByCurso(cursoID uint) ([]models.Certificado, error) {
	var list []models.Certificado
	err := r.db.Where("curso_id = ?", cursoID).Order("id asc").Find(&list).Error
	return list, err
}

func (r *CertificadoRepository) ListElegibilidadesByCurso(cursoID uint) ([]models.CertificadoElegibilidade, error) {
	var list []models.CertificadoElegibilidade
	err := r.db.Where("curso_id = ?", cursoID).Order("id asc").Find(&list).Error
	return list, err
}

// CreateValido cria certificado com status valido e snapshots. Falha se já houver valido.
func (r *CertificadoRepository) CreateValido(userID, cursoID uint, alunoNome, cursoTitulo string) (*models.Certificado, error) {
	now := timeutil.NewDateTime(time.Now())
	c := models.Certificado{
		UserID:      userID,
		CursoID:     cursoID,
		Status:      models.CertificadoStatusValido,
		EmitidoEm:   now,
		AlunoNome:   alunoNome,
		CursoTitulo: cursoTitulo,
	}
	if err := r.db.Create(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// DeleteByID remove um certificado (rollback se PDF falhar na 1ª emissão).
func (r *CertificadoRepository) DeleteByID(id uint) error {
	result := r.db.Delete(&models.Certificado{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Invalidar marca status=invalidado e remove elegibilidade do par (atômico).
func (r *CertificadoRepository) Invalidar(cursoID, certID uint) (*models.Certificado, error) {
	var out *models.Certificado
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var c models.Certificado
		if err := tx.First(&c, certID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if c.CursoID != cursoID {
			return ErrCertificadoNaoPertenceCurso
		}
		if c.Status != models.CertificadoStatusValido {
			return ErrCertificadoJaInvalidado
		}
		c.Status = models.CertificadoStatusInvalidado
		if err := tx.Save(&c).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND curso_id = ?", c.UserID, c.CursoID).
			Delete(&models.CertificadoElegibilidade{}).Error; err != nil {
			return err
		}
		out = &c
		return nil
	})
	return out, err
}
