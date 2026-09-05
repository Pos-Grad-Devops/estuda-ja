package certpdf

import (
	"bytes"
	"fmt"

	certasset "github.com/Pos-Grad-Devops/estuda-ja/backend/assets/certs"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/timeutil"
	"github.com/go-pdf/fpdf"
)

const fontFamily = "DejaVu"

// Generate monta 1 página A4 com título fixo, nome do aluno, curso e data BR (snapshots).
func Generate(alunoNome, cursoTitulo string, emitidoEm timeutil.DateTime) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	pdf.AddUTF8FontFromBytes(fontFamily, "", certasset.DejaVuSansTTF)
	pdf.AddUTF8FontFromBytes(fontFamily, "B", certasset.DejaVuSansTTF)

	pdf.SetFont(fontFamily, "B", 22)
	pdf.CellFormat(0, 14, "Certificado de Conclusão", "", 1, "C", false, 0, "")
	pdf.Ln(12)

	pdf.SetFont(fontFamily, "", 14)
	pdf.MultiCell(0, 8, "Certificamos que", "", "C", false)
	pdf.Ln(4)

	pdf.SetFont(fontFamily, "B", 18)
	pdf.MultiCell(0, 10, alunoNome, "", "C", false)
	pdf.Ln(4)

	pdf.SetFont(fontFamily, "", 14)
	pdf.MultiCell(0, 8, "concluiu o curso", "", "C", false)
	pdf.Ln(4)

	pdf.SetFont(fontFamily, "B", 16)
	pdf.MultiCell(0, 9, cursoTitulo, "", "C", false)
	pdf.Ln(10)

	data := timeutil.FormatDateTime(emitidoEm.Time)
	if data == "" {
		data = timeutil.FormatDate(emitidoEm.Time)
	}
	pdf.SetFont(fontFamily, "", 12)
	pdf.MultiCell(0, 8, fmt.Sprintf("Emitido em %s", data), "", "C", false)

	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("gerar PDF: %w", err)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("gerar PDF: %w", err)
	}
	return buf.Bytes(), nil
}
