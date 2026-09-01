package models

import "github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleProfessor Role = "professor"
	RoleAluno     Role = "aluno"
)

type User struct {
	ID           uint              `json:"id" gorm:"primaryKey"`
	Nome         string            `json:"nome" gorm:"not null"`
	Email        string            `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string            `json:"-" gorm:"not null"`
	Role         Role              `json:"role" gorm:"not null"`
	CreatedAt    timeutil.DateTime `json:"created_at"`
	UpdatedAt    timeutil.DateTime `json:"updated_at"`
}

func (u User) Public() User {
	u.PasswordHash = ""
	return u
}
