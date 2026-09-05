package migrations

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	vodasset "github.com/Pos-Grad-Devops/estuda-ja/backend/assets/vod"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/vodstorage"
	"gorm.io/gorm"
)

const demoVodAssetRel = "assets/vod/demo-aula.mp4"

func seedVodDemo(cfg config.Config) func(*gorm.DB) error {
	return func(tx *gorm.DB) error {
		aula, found, err := findDemoAula(tx)
		if err != nil {
			return err
		}
		if !found {
			log.Println("migration 003: aula demo ausente — seed VOD ignorado")
			return nil
		}

		var count int64
		if err := tx.Model(&models.AulaVod{}).Where("aula_id = ?", aula.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			log.Println("migration 003: VOD demo já existe — skip")
			return nil
		}

		data, err := readDemoVodAsset()
		if err != nil {
			return err
		}
		if !vodstorage.HasMP4Ftyp(data) {
			return fmt.Errorf("migration 003: asset seed não é MP4 válido")
		}

		store, err := vodstorage.NewFromConfig(context.Background(), cfg)
		if err != nil {
			return fmt.Errorf("migration 003: storage VOD: %w", err)
		}

		key := vodstorage.ObjectKey(aula.ID)
		size := int64(len(data))
		if err := store.Put(context.Background(), key, bytes.NewReader(data), size, models.VodContentTypeMP4); err != nil {
			return fmt.Errorf("migration 003: put VOD: %w", err)
		}

		rec := models.AulaVod{
			AulaID:      aula.ID,
			StorageKey:  key,
			ContentType: models.VodContentTypeMP4,
			SizeBytes:   size,
			Status:      models.VodStatusPublicado,
		}
		if err := tx.Create(&rec).Error; err != nil {
			_ = store.Delete(context.Background(), key)
			return err
		}

		log.Println("migration 003: seed VOD demo aplicado (1 gravação publicada na aula de exemplo)")
		return nil
	}
}

func rollbackVodDemo(cfg config.Config) func(*gorm.DB) error {
	return func(tx *gorm.DB) error {
		aula, found, err := findDemoAula(tx)
		if err != nil {
			return err
		}
		if !found {
			return nil
		}

		var rec models.AulaVod
		err = tx.Where("aula_id = ?", aula.ID).First(&rec).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}

		store, err := vodstorage.NewFromConfig(context.Background(), cfg)
		if err == nil {
			_ = store.Delete(context.Background(), rec.StorageKey)
		}
		return tx.Delete(&rec).Error
	}
}

func findDemoAula(tx *gorm.DB) (*models.Aula, bool, error) {
	var curso models.Curso
	err := tx.Where("titulo = ?", demoCursoTitulo).First(&curso).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var aula models.Aula
	err = tx.Where("curso_id = ? AND titulo = ?", curso.ID, demoAulaTitulo).First(&aula).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &aula, true, nil
}

func readDemoVodAsset() ([]byte, error) {
	if len(vodasset.DemoAulaMP4) > 0 {
		return vodasset.DemoAulaMP4, nil
	}
	for _, path := range demoVodAssetCandidates() {
		data, err := os.ReadFile(path)
		if err == nil && len(data) > 0 {
			return data, nil
		}
	}
	return nil, fmt.Errorf("migration 003: asset %s não encontrado (copie no Dockerfile e no repo)", demoVodAssetRel)
}

func demoVodAssetCandidates() []string {
	candidates := []string{
		demoVodAssetRel,
		filepath.Join("/app", demoVodAssetRel),
	}
	if root := findGoModDir(); root != "" {
		candidates = append(candidates, filepath.Join(root, demoVodAssetRel))
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, demoVodAssetRel),
			filepath.Join(wd, "..", "..", "..", demoVodAssetRel),
		)
	}
	return candidates
}

func findGoModDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
