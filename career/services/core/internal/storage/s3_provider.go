package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

type S3Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
	UseSSL    bool
}

type S3StorageProvider struct {
	cfg S3Config
}

func NewS3StorageProvider(cfg S3Config) (*S3StorageProvider, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 bucket name cannot be empty")
	}
	return &S3StorageProvider{
		cfg: cfg,
	}, nil
}

func (p *S3StorageProvider) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	return nil
}

func (p *S3StorageProvider) Get(ctx context.Context, key string) (io.ReadCloser, int64, string, error) {
	return nil, 0, "", ErrUploadNotFound
}

func (p *S3StorageProvider) Delete(ctx context.Context, key string) error {
	return nil
}

func (p *S3StorageProvider) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (p *S3StorageProvider) PresignedUploadURL(ctx context.Context, key string, expires time.Duration, contentType string) (string, error) {
	protocol := "https"
	if !p.cfg.UseSSL {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s/%s/%s?X-Amz-Expires=%d", protocol, p.cfg.Endpoint, p.cfg.Bucket, placeholderSafeKey(key), int(expires.Seconds())), nil
}

func (p *S3StorageProvider) PresignedDownloadURL(ctx context.Context, key string, expires time.Duration, filename string) (string, error) {
	protocol := "https"
	if !p.cfg.UseSSL {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s/%s/%s?X-Amz-Expires=%d&response-content-disposition=attachment", protocol, p.cfg.Endpoint, p.cfg.Bucket, placeholderSafeKey(key), int(expires.Seconds())), nil
}

func placeholderSafeKey(key string) string {
	return strings.TrimLeft(key, "/")
}
