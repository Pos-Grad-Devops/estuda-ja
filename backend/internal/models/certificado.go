package models

import "github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"

const (
	CertificadoStatusValido     = "valido"
	CertificadoStatusInvalidado = "invalidado"
)

// CertificadoElegibilidade: existência da linha = aluno elegível ao certificado do curso.
// UNIQUE (user_id, curso_id). Não cria Certificado.
type CertificadoElegibilidade struct {
	ID        uint              `json:"id" gorm:"primaryKey"`
	UserID    uint              `json:"user_id" gorm:"not null;uniqueIndex:idx_cert_eleg_user_curso"`
	CursoID   uint              `json:"curso_id" gorm:"not null;uniqueIndex:idx_cert_eleg_user_curso"`
	CreatedAt timeutil.DateTime `json:"created_at"`
	UpdatedAt timeutil.DateTime `json:"updated_at"`
}

func (CertificadoElegibilidade) TableName() string {
	return "certificado_elegibilidades"
}

// Certificado: comprovante emitido na 1ª solicitação PDF do aluno elegível.
// No máximo um status=valido por (user_id, curso_id) — índice parcial em Migrate.
type Certificado struct {
	ID          uint              `json:"id" gorm:"primaryKey"`
	UserID      uint              `json:"user_id" gorm:"not null;index"`
	CursoID     uint              `json:"curso_id" gorm:"not null;index"`
	Status      string            `json:"status" gorm:"not null;index"`
	EmitidoEm   timeutil.DateTime `json:"emitido_em" gorm:"not null"`
	AlunoNome   string            `json:"aluno_nome" gorm:"not null"`
	CursoTitulo string            `json:"curso_titulo" gorm:"not null"`
	CreatedAt   timeutil.DateTime `json:"created_at"`
	UpdatedAt   timeutil.DateTime `json:"updated_at"`
}

func (Certificado) TableName() string {
	return "certificados"
}
