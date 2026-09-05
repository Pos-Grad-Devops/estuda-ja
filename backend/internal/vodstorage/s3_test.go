package vodstorage_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/config"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/vodstorage"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

type mockS3Client struct {
	putCalls    int
	getCalls    int
	deleteCalls int
	objects     map[string][]byte
	putErr      error
	getErr      error
	deleteErr   error
}

func (m *mockS3Client) PutObject(_ context.Context, params *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.putCalls++
	if m.putErr != nil {
		return nil, m.putErr
	}
	if m.objects == nil {
		m.objects = map[string][]byte{}
	}
	data, err := io.ReadAll(params.Body)
	if err != nil {
		return nil, err
	}
	m.objects[aws.ToString(params.Key)] = data
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3Client) GetObject(_ context.Context, params *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	m.getCalls++
	if m.getErr != nil {
		return nil, m.getErr
	}
	key := aws.ToString(params.Key)
	data, ok := m.objects[key]
	if !ok {
		return nil, io.EOF
	}
	n := int64(len(data))
	return &s3.GetObjectOutput{
		Body:          io.NopCloser(bytes.NewReader(data)),
		ContentLength: &n,
	}, nil
}

func (m *mockS3Client) DeleteObject(_ context.Context, params *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.deleteCalls++
	if m.deleteErr != nil {
		return nil, m.deleteErr
	}
	delete(m.objects, aws.ToString(params.Key))
	return &s3.DeleteObjectOutput{}, nil
}

type mockPresigner struct {
	calls int
	url   string
	err   error
}

func (m *mockPresigner) PresignGetObject(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	u := m.url
	if u == "" {
		u = "https://example.s3.amazonaws.com/vod/aulas/1/current.mp4?X-Amz-Signature=test"
	}
	return &v4.PresignedHTTPRequest{URL: u}, nil
}

func TestS3PutOpenDeleteWithMock(t *testing.T) {
	client := &mockS3Client{}
	store := vodstorage.NewS3WithClients("test-bucket", client, &mockPresigner{})
	key := vodstorage.ObjectKey(7)
	body := []byte("ftypisomfake-mp4-bytes")

	err := store.Put(context.Background(), key, bytes.NewReader(body), int64(len(body)), "video/mp4")
	require.NoError(t, err)
	require.Equal(t, 1, client.putCalls)

	rc, size, err := store.Open(context.Background(), key)
	require.NoError(t, err)
	defer rc.Close()
	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, int64(len(body)), size)
	require.Equal(t, body, got)

	require.NoError(t, store.Delete(context.Background(), key))
	require.Equal(t, 1, client.deleteCalls)
	_, ok := client.objects[key]
	require.False(t, ok)
}

func TestS3PresignGetWithMock(t *testing.T) {
	presign := &mockPresigner{url: "https://presigned.example/vod.mp4?sig=1"}
	store := vodstorage.NewS3WithClients("test-bucket", &mockS3Client{}, presign)

	url, expires, err := store.PresignGet(context.Background(), vodstorage.ObjectKey(1), 15*time.Minute)
	require.NoError(t, err)
	require.Equal(t, "https://presigned.example/vod.mp4?sig=1", url)
	require.WithinDuration(t, time.Now().Add(15*time.Minute), expires, 5*time.Second)
	require.Equal(t, 1, presign.calls)

	var _ vodstorage.Presigner = store
}

func TestNewFromConfigLocal(t *testing.T) {
	dir := t.TempDir()
	store, err := vodstorage.NewFromConfig(context.Background(), config.Config{
		VODBackend:  "local",
		VODLocalDir: dir,
	})
	require.NoError(t, err)
	require.NotNil(t, store)
	_, isPresigner := store.(vodstorage.Presigner)
	require.False(t, isPresigner)
}

func TestNewFromConfigS3RequiresBucket(t *testing.T) {
	_, err := vodstorage.NewFromConfig(context.Background(), config.Config{
		VODBackend:  "s3",
		VODS3Bucket: "",
		AWSRegion:   "us-east-1",
	})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "VOD_S3_BUCKET"))
}

func TestNewFromConfigInvalidBackend(t *testing.T) {
	_, err := vodstorage.NewFromConfig(context.Background(), config.Config{
		VODBackend: "redis",
	})
	require.Error(t, err)
}

func TestNewS3EmptyBucket(t *testing.T) {
	_, err := vodstorage.NewS3(context.Background(), "  ", "us-east-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "VOD_S3_BUCKET")
}
