package migrations

import (
	"errors"
	"log"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"gorm.io/gorm"
)

// seedCertificadoDemo pré-marca só elegibilidade do aluno seed no curso demo.
// Não cria Certificado (emissão lazy na 1ª solicitação PDF).
func seedCertificadoDemo(cfg config.Config) func(*gorm.DB) error {
	return func(tx *gorm.DB) error {
		aluno, curso, ok, err := findDemoAlunoCurso(tx, cfg.DemoAlunoEmail)
		if err != nil {
			return err
		}
		if !ok {
			log.Println("migration 004: aluno ou curso demo ausente — seed certificado ignorado")
			return nil
		}

		var count int64
		if err := tx.Model(&models.CertificadoElegibilidade{}).
			Where("user_id = ? AND curso_id = ?", aluno.ID, curso.ID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			log.Println("migration 004: elegibilidade demo já existe — skip")
			return nil
		}

		if err := tx.Create(&models.CertificadoElegibilidade{
			UserID:  aluno.ID,
			CursoID: curso.ID,
		}).Error; err != nil {
			return err
		}

		log.Println("migration 004: seed certificado demo aplicado (só elegibilidade aluno×curso; sem PDF pré-emitido)")
		return nil
	}
}

func rollbackCertificadoDemo(cfg config.Config) func(*gorm.DB) error {
	return func(tx *gorm.DB) error {
		aluno, curso, ok, err := findDemoAlunoCurso(tx, cfg.DemoAlunoEmail)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		return tx.Where("user_id = ? AND curso_id = ?", aluno.ID, curso.ID).
			Delete(&models.CertificadoElegibilidade{}).Error
	}
}

func findDemoAlunoCurso(tx *gorm.DB, alunoEmail string) (*models.User, *models.Curso, bool, error) {
	var aluno models.User
	err := tx.Where("email = ? AND role = ?", alunoEmail, models.RoleAluno).First(&aluno).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}

	var curso models.Curso
	err = tx.Where("titulo = ?", demoCursoTitulo).First(&curso).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}
	return &aluno, &curso, true, nil
}
