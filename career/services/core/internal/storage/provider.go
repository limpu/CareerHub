package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StorageProvider interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, int64, string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	PresignedUploadURL(ctx context.Context, key string, expires time.Duration, contentType string) (string, error)
	PresignedDownloadURL(ctx context.Context, key string, expires time.Duration, filename string) (string, error)
}

type LocalStorageProvider struct {
	baseDir    string
	signingKey []byte
	apiBaseURL string
	mu         sync.RWMutex
}

func NewLocalStorageProvider(baseDir string, signingKey []byte, apiBaseURL string) (*LocalStorageProvider, error) {
	if err := os.MkdirAll(baseDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create local storage directory: %w", err)
	}
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:8080"
	}
	return &LocalStorageProvider{
		baseDir:    baseDir,
		signingKey: signingKey,
		apiBaseURL: strings.TrimRight(apiBaseURL, "/"),
	}, nil
}

func (p *LocalStorageProvider) resolvePath(key string) string {
	safeKey := SanitizeFilename(key)
	return filepath.Join(p.baseDir, safeKey)
}

func (p *LocalStorageProvider) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	targetPath := p.resolvePath(key)
	tmpPath := targetPath + ".tmp"

	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create local storage file: %w", err)
	}

	written, err := io.Copy(f, r)
	closeErr := f.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write local storage content: %w", err)
	}
	if closeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close local storage file: %w", closeErr)
	}

	if size > 0 && written != size {
		os.Remove(tmpPath)
		return fmt.Errorf("file size mismatch: expected %d, got %d", size, written)
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to commit local storage file: %w", err)
	}

	metaPath := targetPath + ".meta"
	_ = os.WriteFile(metaPath, []byte(contentType), 0600)

	return nil
}

func (p *LocalStorageProvider) Get(ctx context.Context, key string) (io.ReadCloser, int64, string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	targetPath := p.resolvePath(key)
	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, "", ErrUploadNotFound
		}
		return nil, 0, "", err
	}

	f, err := os.Open(targetPath)
	if err != nil {
		return nil, 0, "", err
	}

	contentType := "application/octet-stream"
	metaPath := targetPath + ".meta"
	if metaBytes, err := os.ReadFile(metaPath); err == nil && len(metaBytes) > 0 {
		contentType = string(metaBytes)
	}

	return f, info.Size(), contentType, nil
}

func (p *LocalStorageProvider) Delete(ctx context.Context, key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	targetPath := p.resolvePath(key)
	_ = os.Remove(targetPath + ".meta")
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (p *LocalStorageProvider) Exists(ctx context.Context, key string) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	targetPath := p.resolvePath(key)
	_, err := os.Stat(targetPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (p *LocalStorageProvider) GenerateSignedToken(action, key string, expires time.Duration) string {
	expUnix := time.Now().Add(expires).Unix()
	payload := fmt.Sprintf("%s:%s:%d", action, key, expUnix)

	h := hmac.New(sha256.New, p.signingKey)
	h.Write([]byte(payload))
	sig := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s:%s", payload, sig)
}

func (p *LocalStorageProvider) ValidateSignedToken(tokenStr, expectedAction, expectedKey string) error {
	parts := strings.Split(tokenStr, ":")
	if len(parts) != 4 {
		return ErrInvalidStorageToken
	}

	action := parts[0]
	key := parts[1]
	expStr := parts[2]
	givenSig := parts[3]

	if expectedAction != "" && action != expectedAction {
		return ErrInvalidStorageToken
	}
	if expectedKey != "" && key != expectedKey {
		return ErrInvalidStorageToken
	}

	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || time.Now().Unix() > expUnix {
		return ErrInvalidStorageToken
	}

	payload := fmt.Sprintf("%s:%s:%d", action, key, expUnix)
	h := hmac.New(sha256.New, p.signingKey)
	h.Write([]byte(payload))
	expectedSig := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(givenSig), []byte(expectedSig)) {
		return ErrInvalidStorageToken
	}

	return nil
}

func (p *LocalStorageProvider) PresignedUploadURL(ctx context.Context, key string, expires time.Duration, contentType string) (string, error) {
	token := p.GenerateSignedToken("upload", key, expires)
	return fmt.Sprintf("%s/api/v1/storage/upload?token=%s", p.apiBaseURL, token), nil
}

func (p *LocalStorageProvider) PresignedDownloadURL(ctx context.Context, key string, expires time.Duration, filename string) (string, error) {
	token := p.GenerateSignedToken("download", key, expires)
	safeFilename := SanitizeFilename(filename)
	return fmt.Sprintf("%s/api/v1/storage/download?token=%s&filename=%s", p.apiBaseURL, token, safeFilename), nil
}
