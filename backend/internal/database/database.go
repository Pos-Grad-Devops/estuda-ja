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
	return db.AutoMigrate(
		&models.User{},
		&models.Curso{},
		&models.Aula{},
		&models.Aluno{},
		&models.AulaVod{},
	)
}

func RunMigrations(db *gorm.DB, cfg config.Config) error {
	if err := Migrate(db); err != nil {
		return err
	}
	return migrations.Run(db, cfg)
}
