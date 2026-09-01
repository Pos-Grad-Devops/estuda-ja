package migrations

import (
	"log"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/auth"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"gorm.io/gorm"
)

func seedAdmin(cfg config.Config) func(*gorm.DB) error {
	return func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			log.Println("migration 001: admin já existe, pulando seed")
			return nil
		}

		hash, err := auth.HashPassword(cfg.AdminPassword)
		if err != nil {
			return err
		}

		admin := models.User{
			Nome:         "Administrador",
			Email:        cfg.AdminEmail,
			PasswordHash: hash,
			Role:         models.RoleAdmin,
		}
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}

		log.Printf("migration 001: admin inicial criado (%s)", cfg.AdminEmail)
		return nil
	}
}

func rollbackAdmin(cfg config.Config) func(*gorm.DB) error {
	return func(tx *gorm.DB) error {
		return tx.Where("email = ? AND role = ?", cfg.AdminEmail, models.RoleAdmin).
			Delete(&models.User{}).Error
	}
}
