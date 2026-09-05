package database

import (
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/database/migrations"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.User{},
		&models.Curso{},
		&models.Aula{},
		&models.Aluno{},
		&models.AulaVod{},
		&models.AulaLive{},
	); err != nil {
		return err
	}
	// Invariante ≤1 live ao_vivo (Postgres/SQLite com índice parcial).
	_ = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_aula_lives_one_ao_vivo ON aula_lives (status) WHERE status = 'ao_vivo'`).Error
	return nil
}

func RunMigrations(db *gorm.DB, cfg config.Config) error {
	if err := Migrate(db); err != nil {
		return err
	}
	return migrations.Run(db, cfg)
}
