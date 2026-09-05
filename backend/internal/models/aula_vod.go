package models

import "github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"

const (
	VodStatusPublicado = "publicado"
	VodStatusRascunho  = "rascunho" // P2
	VodContentTypeMP4  = "video/mp4"
	VodMaxSizeBytes    = 52_428_800 // 50 MiB
)

// AulaVod é o único conteúdo gravado vigente de uma aula (FR-004).
type AulaVod struct {
	ID          uint              `json:"id" gorm:"primaryKey"`
	AulaID      uint              `json:"aula_id" gorm:"uniqueIndex;not null"`
	Aula        *Aula             `json:"aula,omitempty" gorm:"foreignKey:AulaID"`
	StorageKey  string            `json:"-" gorm:"not null"`
	ContentType string            `json:"content_type" gorm:"not null"`
	SizeBytes   int64             `json:"size_bytes" gorm:"not null"`
	Status      string            `json:"status" gorm:"not null;default:publicado"`
	Titulo      string            `json:"titulo,omitempty"`
	DuracaoSeg  *int              `json:"duracao_segundos,omitempty" gorm:"column:duracao_segundos"`
	CreatedAt   timeutil.DateTime `json:"created_at"`
	UpdatedAt   timeutil.DateTime `json:"updated_at"`
}
