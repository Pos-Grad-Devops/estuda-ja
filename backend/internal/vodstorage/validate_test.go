package vodstorage_test

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"testing"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/vodstorage"
	"github.com/stretchr/testify/require"
)

func TestHasMP4Ftyp(t *testing.T) {
	ok := []byte{0, 0, 0, 0x18, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm'}
	require.True(t, vodstorage.HasMP4Ftyp(ok))
	require.False(t, vodstorage.HasMP4Ftyp([]byte("notmp4!!")))
	require.False(t, vodstorage.HasMP4Ftyp([]byte("short")))
}

func TestValidateUploadRejectsOversize(t *testing.T) {
	h := &multipart.FileHeader{
		Filename: "aula.mp4",
		Size:     models.VodMaxSizeBytes + 1,
		Header:   textproto.MIMEHeader{"Content-Type": []string{"video/mp4"}},
	}
	err := vodstorage.ValidateUpload(h, func() (multipart.File, error) {
		t.Fatal("não deve abrir arquivo acima do limite")
		return nil, nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "50 MB")
}

func TestValidateUploadRejectsNonMP4(t *testing.T) {
	h := &multipart.FileHeader{
		Filename: "aula.txt",
		Size:     12,
		Header:   textproto.MIMEHeader{"Content-Type": []string{"text/plain"}},
	}
	err := vodstorage.ValidateUpload(h, func() (multipart.File, error) {
		return &bufferFile{Reader: bytes.NewReader([]byte("hello world!"))}, nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "MP4")
}

type bufferFile struct {
	*bytes.Reader
}

func (b *bufferFile) Close() error { return nil }
