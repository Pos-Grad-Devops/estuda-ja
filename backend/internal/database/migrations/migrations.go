package migrations

import (
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func Run(db *gorm.DB, cfg config.Config) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, All(cfg))
	return m.Migrate()
}

func All(cfg config.Config) []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID:       "001_seed_admin",
			Migrate:  seedAdmin(cfg),
			Rollback: rollbackAdmin(cfg),
		},
		{
			ID:       "002_seed_demo",
			Migrate:  seedDemo(cfg),
			Rollback: rollbackDemo(cfg),
		},
		{
			ID:       "003_seed_vod_demo",
			Migrate:  seedVodDemo(cfg),
			Rollback: rollbackVodDemo(cfg),
		},
	}
}
