package models

import "github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"

type Curso struct {
	ID        uint              `json:"id" gorm:"primaryKey"`
	Titulo    string            `json:"titulo" gorm:"not null"`
	Descricao string            `json:"descricao"`
	CreatedAt timeutil.DateTime `json:"created_at"`
	UpdatedAt timeutil.DateTime `json:"updated_at"`
	Aulas     []Aula            `json:"aulas,omitempty"`
}

type Aula struct {
	ID         uint              `json:"id" gorm:"primaryKey"`
	CursoID    uint              `json:"curso_id" gorm:"not null;index"`
	Curso      *Curso            `json:"curso,omitempty" gorm:"foreignKey:CursoID"`
	Titulo     string            `json:"titulo" gorm:"not null"`
	Descricao  string            `json:"descricao"`
	AgendadaEm timeutil.DateTime `json:"agendada_em"`
	Status     string            `json:"status" gorm:"default:agendada"`
	CreatedAt  timeutil.DateTime `json:"created_at"`
	UpdatedAt  timeutil.DateTime `json:"updated_at"`
}

const (
	AulaStatusAgendada  = "agendada"
	AulaStatusAoVivo    = "ao_vivo"
	AulaStatusEncerrada = "encerrada"
)

type Aluno struct {
	ID        uint              `json:"id" gorm:"primaryKey"`
	Nome      string            `json:"nome" gorm:"not null"`
	Email     string            `json:"email" gorm:"uniqueIndex;not null"`
	CreatedAt timeutil.DateTime `json:"created_at"`
	UpdatedAt timeutil.DateTime `json:"updated_at"`
}
