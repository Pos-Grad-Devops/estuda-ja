package certpdf_test

import (
	"testing"
	"time"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/certpdf"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"
	"github.com/stretchr/testify/require"
)

func TestGeneratePDFComAcentos(t *testing.T) {
	emitido := timeutil.NewDateTime(time.Date(2026, 9, 5, 17, 30, 0, 0, time.Local))
	pdf, err := certpdf.Generate("José da Conceição", "Aula Magna — Direito Constitucional", emitido)
	require.NoError(t, err)
	require.Greater(t, len(pdf), 100)
	require.Equal(t, "%PDF", string(pdf[:4]))
}
