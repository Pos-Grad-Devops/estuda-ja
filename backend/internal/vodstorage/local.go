package vodstorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Local grava objetos sob root com o mesmo path relativo das keys S3.
type Local struct {
	root string
}

func NewLocal(root string) (*Local, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, fmt.Errorf("diretório VOD local não configurado")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("diretório VOD inválido: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("não foi possível criar diretório VOD: %w", err)
	}
	return &Local{root: abs}, nil
}

func (l *Local) resolve(key string) (string, error) {
	key = strings.TrimPrefix(filepath.ToSlash(key), "/")
	if key == "" || strings.Contains(key, "..") {
		return "", fmt.Errorf("chave de storage inválida")
	}
	full := filepath.Join(l.root, filepath.FromSlash(key))
	abs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	rootWithSep := l.root + string(os.PathSeparator)
	if abs != l.root && !strings.HasPrefix(abs, rootWithSep) {
		return "", fmt.Errorf("chave de storage fora do diretório VOD")
	}
	return abs, nil
}

func (l *Local) Put(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	path, err := l.resolve(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "upload-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		_ = tmp.Close()
		if !committed {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := io.Copy(tmp, r); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	committed = true
	return nil
}

func (l *Local) Open(_ context.Context, key string) (io.ReadCloser, int64, error) {
	path, err := l.resolve(key)
	if err != nil {
		return nil, 0, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, fmt.Errorf("objeto VOD não encontrado")
		}
		return nil, 0, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, err
	}
	return f, info.Size(), nil
}

func (l *Local) Delete(_ context.Context, key string) error {
	path, err := l.resolve(key)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
