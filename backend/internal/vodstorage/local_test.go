package vodstorage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/vodstorage"
	"github.com/stretchr/testify/require"
)

func TestLocalPutOpenDelete(t *testing.T) {
	dir := t.TempDir()
	store, err := vodstorage.NewLocal(dir)
	require.NoError(t, err)

	key := vodstorage.ObjectKey(42)
	payload := []byte{0, 0, 0, 0x18, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 1, 2, 3}
	require.NoError(t, store.Put(context.Background(), key, bytes.NewReader(payload), int64(len(payload)), "video/mp4"))

	full := filepath.Join(dir, filepath.FromSlash(key))
	_, err = os.Stat(full)
	require.NoError(t, err)

	rc, size, err := store.Open(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, int64(len(payload)), size)
	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.NoError(t, rc.Close())
	require.Equal(t, payload, got)

	require.NoError(t, store.Delete(context.Background(), key))
	_, err = os.Stat(full)
	require.True(t, os.IsNotExist(err))
}
