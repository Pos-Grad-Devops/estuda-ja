package vodstorage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// s3Client cobre Put/Get/Delete usados pelo wrapper (mockável em testes).
type s3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// s3Presigner cobre PresignGetObject (mockável em testes).
type s3Presigner interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// S3 grava objetos no bucket efêmero da demo (task role IAM).
type S3 struct {
	bucket   string
	client   s3Client
	presign  s3Presigner
}

// NewS3 carrega o AWS SDK v2 (credenciais da task role / default chain) e falha se bucket vazio.
func NewS3(ctx context.Context, bucket, region string) (*S3, error) {
	bucket = strings.TrimSpace(bucket)
	if bucket == "" {
		return nil, fmt.Errorf("VOD_S3_BUCKET é obrigatório quando VOD_BACKEND=s3")
	}
	region = strings.TrimSpace(region)
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("falha ao carregar configuração AWS: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	return NewS3WithClients(bucket, client, s3.NewPresignClient(client)), nil
}

// NewS3WithClients permite injetar clientes (testes unitários sem conta AWS).
func NewS3WithClients(bucket string, client s3Client, presign s3Presigner) *S3 {
	return &S3{bucket: bucket, client: client, presign: presign}
}

func (s *S3) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	key = normalizeKey(key)
	if key == "" {
		return fmt.Errorf("chave de storage inválida")
	}
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   r,
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	if size > 0 {
		input.ContentLength = aws.Int64(size)
	}
	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("falha ao gravar objeto VOD no S3: %w", err)
	}
	return nil
}

func (s *S3) Open(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	key = normalizeKey(key)
	if key == "" {
		return nil, 0, fmt.Errorf("chave de storage inválida")
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("objeto VOD não encontrado: %w", err)
	}
	var size int64
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	return out.Body, size, nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	key = normalizeKey(key)
	if key == "" {
		return fmt.Errorf("chave de storage inválida")
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("falha ao remover objeto VOD no S3: %w", err)
	}
	return nil
}

// PresignGet implementa Presigner — URL GetObject com TTL (~15 min na demo).
func (s *S3) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, time.Time, error) {
	key = normalizeKey(key)
	if key == "" {
		return "", time.Time{}, fmt.Errorf("chave de storage inválida")
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	expiresAt := time.Now().Add(ttl)
	out, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("falha ao gerar URL de reprodução: %w", err)
	}
	return out.URL, expiresAt, nil
}

func normalizeKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.TrimPrefix(key, "/")
	if key == "" || strings.Contains(key, "..") {
		return ""
	}
	return key
}
