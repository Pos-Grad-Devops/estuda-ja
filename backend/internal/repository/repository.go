package repository

import (
	"errors"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("registro não encontrado")

type Repositories struct {
	Cursos *CursoRepository
	Aulas  *AulaRepository
	Alunos *AlunoRepository
	Users  *UserRepository
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		Cursos: &CursoRepository{db: db},
		Aulas:  &AulaRepository{db: db},
		Alunos: &AlunoRepository{db: db},
		Users:  &UserRepository{db: db},
	}
}
