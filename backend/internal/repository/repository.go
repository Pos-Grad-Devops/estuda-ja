package repository

import (
	"errors"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("registro não encontrado")

type Repositories struct {
	Cursos       *CursoRepository
	Aulas        *AulaRepository
	Alunos       *AlunoRepository
	Users        *UserRepository
	Vod          *VodRepository
	Live         *LiveRepository
	Certificados *CertificadoRepository
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		Cursos:       &CursoRepository{db: db},
		Aulas:        &AulaRepository{db: db},
		Alunos:       &AlunoRepository{db: db},
		Users:        &UserRepository{db: db},
		Vod:          &VodRepository{db: db},
		Live:         &LiveRepository{db: db},
		Certificados: &CertificadoRepository{db: db},
	}
}
