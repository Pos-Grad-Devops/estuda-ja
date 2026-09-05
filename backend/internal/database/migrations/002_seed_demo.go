package migrations

import (
	"log"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/auth"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"
	"gorm.io/gorm"
)

const (
	demoCursoTitulo = "Aula Magna — Direito Constitucional"
	demoAulaTitulo  = "Abertura da aula magna"
	demoAulaQuando  = "15/09/2026 19:00"
)

func seedDemo(cfg config.Config) func(*gorm.DB) error {
	return func(tx *gorm.DB) error {
		if err := ensureUser(tx, "Professor Demo", cfg.DemoProfessorEmail, cfg.DemoProfessorPassword, models.RoleProfessor); err != nil {
			return err
		}
		if err := ensureUser(tx, "Aluno Demo", cfg.DemoAlunoEmail, cfg.DemoAlunoPassword, models.RoleAluno); err != nil {
			return err
		}

		curso, err := ensureCurso(tx)
		if err != nil {
			return err
		}
		if err := ensureAula(tx, curso.ID); err != nil {
			return err
		}

		log.Println("migration 002: seed demo aplicado (professor, aluno, 1 curso, 1 aula)")
		return nil
	}
}

func rollbackDemo(cfg config.Config) func(*gorm.DB) error {
	return func(tx *gorm.DB) error {
		if err := tx.Where("titulo = ?", demoAulaTitulo).Delete(&models.Aula{}).Error; err != nil {
			return err
		}
		if err := tx.Where("titulo = ?", demoCursoTitulo).Delete(&models.Curso{}).Error; err != nil {
			return err
		}
		return tx.Where("email IN ?", []string{cfg.DemoProfessorEmail, cfg.DemoAlunoEmail}).
			Delete(&models.User{}).Error
	}
}

func ensureUser(tx *gorm.DB, nome, email, password string, role models.Role) error {
	var count int64
	if err := tx.Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	return tx.Create(&models.User{
		Nome:         nome,
		Email:        email,
		PasswordHash: hash,
		Role:         role,
	}).Error
}

func ensureCurso(tx *gorm.DB) (*models.Curso, error) {
	var count int64
	if err := tx.Model(&models.Curso{}).Where("titulo = ?", demoCursoTitulo).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		var curso models.Curso
		if err := tx.Where("titulo = ?", demoCursoTitulo).First(&curso).Error; err != nil {
			return nil, err
		}
		return &curso, nil
	}

	curso := models.Curso{
		Titulo:    demoCursoTitulo,
		Descricao: "Curso de exemplo da demo (seed 002). Recriado a cada banco vazio.",
	}
	if err := tx.Create(&curso).Error; err != nil {
		return nil, err
	}
	return &curso, nil
}

func ensureAula(tx *gorm.DB, cursoID uint) error {
	var count int64
	if err := tx.Model(&models.Aula{}).Where("curso_id = ? AND titulo = ?", cursoID, demoAulaTitulo).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	agendada, err := timeutil.ParseDateTime(demoAulaQuando)
	if err != nil {
		return err
	}

	return tx.Create(&models.Aula{
		CursoID:    cursoID,
		Titulo:     demoAulaTitulo,
		Descricao:  "Aula de exemplo ligada ao curso demo.",
		AgendadaEm: timeutil.NewDateTime(agendada),
		Status:     models.AulaStatusAgendada,
	}).Error
}
