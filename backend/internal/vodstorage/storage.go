package vodstorage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Storage abstrai bytes VOD (disco local ou S3).
type Storage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Open(ctx context.Context, key string) (io.ReadCloser, int64, error)
	Delete(ctx context.Context, key string) error
}

// Presigner gera URL temporária de leitura (S3 GetObject presigned).
// Implementações locais não precisam desta interface — o handler usa content token.
type Presigner interface {
	PresignGet(ctx context.Context, key string, ttl time.Duration) (url string, expiresAt time.Time, err error)
}

// ObjectKey retorna a key canônica do VOD vigente da aula.
func ObjectKey(aulaID uint) string {
	return fmt.Sprintf("vod/aulas/%d/current.mp4", aulaID)
}
