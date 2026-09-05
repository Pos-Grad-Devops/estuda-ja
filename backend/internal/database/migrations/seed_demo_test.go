package migrations

import (
	"testing"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/auth"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func demoTestConfig() config.Config {
	return config.Config{
		AdminEmail:            "admin@estudaja.com",
		AdminPassword:         "admin123",
		DemoProfessorEmail:    "professor@estudaja.com",
		DemoProfessorPassword: "professor123",
		DemoAlunoEmail:        "aluno@estudaja.com",
		DemoAlunoPassword:     "aluno123",
	}
}

func setupMigratedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Curso{}, &models.Aula{}, &models.Aluno{}))
	return db
}

func countModel(t *testing.T, db *gorm.DB, model any) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(model).Count(&n).Error)
	return n
}

func TestSeedDemoCreatesAdminProfessorAlunoCursoAula(t *testing.T) {
	db := setupMigratedDB(t)
	cfg := demoTestConfig()

	require.NoError(t, Run(db, cfg))

	require.Equal(t, int64(3), countModel(t, db, &models.User{}))
	require.Equal(t, int64(1), countModel(t, db, &models.Curso{}))
	require.Equal(t, int64(1), countModel(t, db, &models.Aula{}))

	assertUser(t, db, cfg.AdminEmail, models.RoleAdmin, cfg.AdminPassword)
	assertUser(t, db, cfg.DemoProfessorEmail, models.RoleProfessor, cfg.DemoProfessorPassword)
	assertUser(t, db, cfg.DemoAlunoEmail, models.RoleAluno, cfg.DemoAlunoPassword)

	var curso models.Curso
	require.NoError(t, db.Where("titulo = ?", demoCursoTitulo).First(&curso).Error)

	var aula models.Aula
	require.NoError(t, db.Where("curso_id = ? AND titulo = ?", curso.ID, demoAulaTitulo).First(&aula).Error)
	require.Equal(t, models.AulaStatusAgendada, aula.Status)
}

func TestSeedDemoIdempotentOnRerun(t *testing.T) {
	db := setupMigratedDB(t)
	cfg := demoTestConfig()

	require.NoError(t, Run(db, cfg))
	require.NoError(t, Run(db, cfg))
	require.NoError(t, seedDemo(cfg)(db))

	require.Equal(t, int64(3), countModel(t, db, &models.User{}))
	require.Equal(t, int64(1), countModel(t, db, &models.Curso{}))
	require.Equal(t, int64(1), countModel(t, db, &models.Aula{}))
}

func assertUser(t *testing.T, db *gorm.DB, email string, role models.Role, password string) {
	t.Helper()
	var user models.User
	require.NoError(t, db.Where("email = ?", email).First(&user).Error)
	require.Equal(t, role, user.Role)
	require.True(t, auth.CheckPassword(user.PasswordHash, password), email)
}
