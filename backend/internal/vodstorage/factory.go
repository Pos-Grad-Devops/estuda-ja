package vodstorage

import (
	"context"
	"fmt"
	"strings"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
)

// NewFromConfig seleciona local ou S3 conforme VOD_BACKEND (fail-fast se s3 sem bucket).
func NewFromConfig(ctx context.Context, cfg config.Config) (Storage, error) {
	backend := strings.ToLower(strings.TrimSpace(cfg.VODBackend))
	if backend == "" {
		backend = "local"
	}
	switch backend {
	case "local":
		dir := cfg.VODLocalDir
		if dir == "" {
			dir = "./data/vod"
		}
		return NewLocal(dir)
	case "s3":
		return NewS3(ctx, cfg.VODS3Bucket, cfg.AWSRegion)
	default:
		return nil, fmt.Errorf("VOD_BACKEND inválido (use local ou s3)")
	}
}
