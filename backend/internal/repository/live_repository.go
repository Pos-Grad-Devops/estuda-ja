package repository

import (
	"errors"
	"time"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"
	"gorm.io/gorm"
)

var (
	// ErrLiveAlreadyActive — outra aula já está ao_vivo.
	ErrLiveAlreadyActive = errors.New("já existe uma transmissão ao vivo")
	// ErrLiveNotActive — esta aula não está ao_vivo (stop inválido).
	ErrLiveNotActive = errors.New("esta aula não está ao vivo")
	// ErrLiveNotScheduled — cancelar exige status agendada.
	ErrLiveNotScheduled = errors.New("esta aula não tem transmissão agendada")
	// ErrLiveScheduleConflict — não dá para agendar enquanto ao_vivo.
	ErrLiveScheduleConflict = errors.New("não é possível agendar enquanto a transmissão está ao vivo")
)

type LiveRepository struct {
	db *gorm.DB
}

func (r *LiveRepository) GetByAulaID(aulaID uint) (*models.AulaLive, error) {
	var live models.AulaLive
	err := r.db.Where("aula_id = ?", aulaID).First(&live).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &live, err
}

func (r *LiveRepository) GetAtiva() (*models.AulaLive, error) {
	var live models.AulaLive
	err := r.db.Where("status = ?", models.LiveStatusAoVivo).First(&live).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &live, err
}

// Start marca a aula como ao_vivo (upsert). Idempotente se já ao_vivo nesta aula.
// Falha com ErrLiveAlreadyActive se outra aula estiver ao_vivo.
func (r *LiveRepository) Start(aulaID uint) (*models.AulaLive, error) {
	var out *models.AulaLive
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var ativa models.AulaLive
		err := tx.Where("status = ?", models.LiveStatusAoVivo).First(&ativa).Error
		if err == nil && ativa.AulaID != aulaID {
			return ErrLiveAlreadyActive
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := timeutil.NewDateTime(time.Now())
		var existing models.AulaLive
		err = tx.Where("aula_id = ?", aulaID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			live := models.AulaLive{
				AulaID:      aulaID,
				Status:      models.LiveStatusAoVivo,
				IniciadaEm:  &now,
				EncerradaEm: nil,
			}
			if err := tx.Create(&live).Error; err != nil {
				return err
			}
			out = &live
			return nil
		}
		if err != nil {
			return err
		}
		if existing.Status == models.LiveStatusAoVivo {
			out = &existing
			return nil
		}
		existing.Status = models.LiveStatusAoVivo
		existing.IniciadaEm = &now
		existing.EncerradaEm = nil
		if err := tx.Save(&existing).Error; err != nil {
			return err
		}
		out = &existing
		return nil
	})
	return out, err
}

// Schedule marca a live como agendada (upsert). Idempotente se já agendada nesta aula.
// Não permite se esta aula estiver ao_vivo.
func (r *LiveRepository) Schedule(aulaID uint) (*models.AulaLive, error) {
	var out *models.AulaLive
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existing models.AulaLive
		err := tx.Where("aula_id = ?", aulaID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			live := models.AulaLive{
				AulaID:      aulaID,
				Status:      models.LiveStatusAgendada,
				IniciadaEm:  nil,
				EncerradaEm: nil,
			}
			if err := tx.Create(&live).Error; err != nil {
				return err
			}
			out = &live
			return nil
		}
		if err != nil {
			return err
		}
		if existing.Status == models.LiveStatusAoVivo {
			return ErrLiveScheduleConflict
		}
		if existing.Status == models.LiveStatusAgendada {
			out = &existing
			return nil
		}
		existing.Status = models.LiveStatusAgendada
		existing.IniciadaEm = nil
		existing.EncerradaEm = nil
		if err := tx.Save(&existing).Error; err != nil {
			return err
		}
		out = &existing
		return nil
	})
	return out, err
}

// Cancel remove a live agendada (volta a inativa / sem linha).
func (r *LiveRepository) Cancel(aulaID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing models.AulaLive
		err := tx.Where("aula_id = ?", aulaID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLiveNotScheduled
		}
		if err != nil {
			return err
		}
		if existing.Status != models.LiveStatusAgendada {
			return ErrLiveNotScheduled
		}
		return tx.Delete(&existing).Error
	})
}

// Stop encerra a live desta aula. Erro se não estiver ao_vivo.
func (r *LiveRepository) Stop(aulaID uint) (*models.AulaLive, error) {
	var out *models.AulaLive
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existing models.AulaLive
		err := tx.Where("aula_id = ?", aulaID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLiveNotActive
		}
		if err != nil {
			return err
		}
		if existing.Status != models.LiveStatusAoVivo {
			return ErrLiveNotActive
		}
		now := timeutil.NewDateTime(time.Now())
		existing.Status = models.LiveStatusEncerrada
		existing.EncerradaEm = &now
		if err := tx.Save(&existing).Error; err != nil {
			return err
		}
		out = &existing
		return nil
	})
	return out, err
}

func (r *LiveRepository) DeleteByAulaID(aulaID uint) error {
	result := r.db.Where("aula_id = ?", aulaID).Delete(&models.AulaLive{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
