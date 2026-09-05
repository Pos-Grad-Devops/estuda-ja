package vodstorage

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
)

// ValidateUpload confere tamanho ≤ 50 MB, MIME/extensão MP4 e magic ftyp.
// Retorna mensagem de erro em português pronta para 400.
func ValidateUpload(header *multipart.FileHeader, open func() (multipart.File, error)) error {
	if header == nil {
		return fmt.Errorf("arquivo é obrigatório")
	}
	if header.Size > models.VodMaxSizeBytes {
		return fmt.Errorf("arquivo excede o limite de 50 MB")
	}

	ct := strings.TrimSpace(strings.ToLower(header.Header.Get("Content-Type")))
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ct != "" && ct != "application/octet-stream" && ct != models.VodContentTypeMP4 {
		return fmt.Errorf("apenas arquivos MP4 são permitidos")
	}
	if ext != "" && ext != ".mp4" {
		return fmt.Errorf("apenas arquivos MP4 são permitidos")
	}
	if ct == "" && ext == "" {
		return fmt.Errorf("apenas arquivos MP4 são permitidos")
	}

	f, err := open()
	if err != nil {
		return fmt.Errorf("não foi possível ler o arquivo")
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return fmt.Errorf("não foi possível ler o arquivo")
	}
	if n < 8 {
		return fmt.Errorf("apenas arquivos MP4 são permitidos")
	}
	if !HasMP4Ftyp(buf[:n]) {
		return fmt.Errorf("apenas arquivos MP4 são permitidos")
	}
	detected := http.DetectContentType(buf[:n])
	if detected != models.VodContentTypeMP4 && !strings.HasPrefix(detected, "video/") {
		// DetectContentType frequentemente retorna application/octet-stream para MP4 curto;
		// magic ftyp já validado acima.
		if detected != "application/octet-stream" {
			return fmt.Errorf("apenas arquivos MP4 são permitidos")
		}
	}
	return nil
}

// HasMP4Ftyp verifica box ISO BMFF (ftyp nos bytes 4–7).
func HasMP4Ftyp(header []byte) bool {
	if len(header) < 8 {
		return false
	}
	return string(header[4:8]) == "ftyp"
}
