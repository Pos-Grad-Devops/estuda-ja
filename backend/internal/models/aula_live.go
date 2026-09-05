package models

import "github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"

const (
	LiveStatusAgendada  = "agendada"
	LiveStatusAoVivo    = "ao_vivo"
	LiveStatusEncerrada = "encerrada"
	// LiveStatusInativa é só resposta de API (ausência de linha); não persistir.
	LiveStatusInativa = "inativa"
)

// AulaLive associa uma aula à sessão de streaming (canal compartilhado).
// Ausência de linha = live inativa. Status: agendada | ao_vivo | encerrada.
type AulaLive struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	AulaID      uint               `json:"aula_id" gorm:"uniqueIndex;not null"`
	Aula        *Aula              `json:"aula,omitempty" gorm:"foreignKey:AulaID"`
	Status      string             `json:"status" gorm:"not null;index"`
	IniciadaEm  *timeutil.DateTime `json:"iniciada_em"`
	EncerradaEm *timeutil.DateTime `json:"encerrada_em"`
	CreatedAt   timeutil.DateTime  `json:"created_at"`
	UpdatedAt   timeutil.DateTime  `json:"updated_at"`
}
