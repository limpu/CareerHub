package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
)

type UploadRepository interface {
	Create(ctx context.Context, up *UploadRecord) error
	GetByID(ctx context.Context, id string) (*UploadRecord, error)
	UpdateQuarantineStatus(ctx context.Context, id string, status QuarantineStatus, reason string) error
	Delete(ctx context.Context, id string) error
	ListByUserID(ctx context.Context, userID string) ([]*UploadRecord, error)
}

type MemoryUploadRepository struct {
	mu      sync.RWMutex
	records map[string]*UploadRecord
}

func NewMemoryUploadRepository() *MemoryUploadRepository {
	return &MemoryUploadRepository{
		records: make(map[string]*UploadRecord),
	}
}

func (r *MemoryUploadRepository) Create(ctx context.Context, up *UploadRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[up.ID] = up
	return nil
}

func (r *MemoryUploadRepository) GetByID(ctx context.Context, id string) (*UploadRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	up, ok := r.records[id]
	if !ok {
		return nil, ErrUploadNotFound
	}
	return up, nil
}

func (r *MemoryUploadRepository) UpdateQuarantineStatus(ctx context.Context, id string, status QuarantineStatus, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	up, ok := r.records[id]
	if !ok {
		return ErrUploadNotFound
	}
	up.QuarantineStatus = status
	up.RejectionReason = reason
	return nil
}

func (r *MemoryUploadRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.records, id)
	return nil
}

func (r *MemoryUploadRepository) ListByUserID(ctx context.Context, userID string) ([]*UploadRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*UploadRecord
	for _, up := range r.records {
		if up.UserID == userID {
			list = append(list, up)
		}
	}
	return list, nil
}

type UploadService struct {
	repo     UploadRepository
	provider StorageProvider
}

func NewUploadService(repo UploadRepository, provider StorageProvider) *UploadService {
	return &UploadService{
		repo:     repo,
		provider: provider,
	}
}

func (s *UploadService) RequestUpload(ctx context.Context, userID string, req UploadRequest) (*UploadResponse, error) {
	if req.Size > DefaultMaxUploadSize {
		return nil, ErrFileTooLarge
	}

	safeName := SanitizeFilename(req.Filename)
	uploadID := uuid.New().String()
	storageKey := fmt.Sprintf("uploads/%s/%s-%s", userID, uploadID, safeName)

	rec := &UploadRecord{
		ID:               uploadID,
		UserID:           userID,
		OriginalFilename: safeName,
		StorageKey:       storageKey,
		MimeType:         req.ContentType,
		FileSize:         req.Size,
		QuarantineStatus: QuarantineStatusQuarantined,
		CreatedAt:        time.Now(),
	}

	if err := s.repo.Create(ctx, rec); err != nil {
		return nil, fmt.Errorf("failed to create upload record: %w", err)
	}

	expires := 15 * time.Minute
	uploadURL, err := s.provider.PresignedUploadURL(ctx, storageKey, expires, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned upload url: %w", err)
	}

	return &UploadResponse{
		UploadID:   uploadID,
		StorageKey: storageKey,
		UploadURL:  uploadURL,
		ExpiresAt:  time.Now().Add(expires),
	}, nil
}

func (s *UploadService) ProcessAndCommitUpload(ctx context.Context, uploadID string, r io.Reader) (*UploadRecord, error) {
	rec, err := s.repo.GetByID(ctx, uploadID)
	if err != nil {
		return nil, err
	}

	_ = s.repo.UpdateQuarantineStatus(ctx, uploadID, QuarantineStatusScanning, "")

	buf, status, reason, err := InspectStream(r)
	if err != nil || status == QuarantineStatusRejected {
		_ = s.repo.UpdateQuarantineStatus(ctx, uploadID, QuarantineStatusRejected, reason)
		rec.QuarantineStatus = QuarantineStatusRejected
		rec.RejectionReason = reason
		return rec, ErrEncryptedOrMacroContent
	}

	mimeType, err := ValidateFileHeader(rec.OriginalFilename, buf, int64(len(buf)))
	if err != nil {
		reason := err.Error()
		_ = s.repo.UpdateQuarantineStatus(ctx, uploadID, QuarantineStatusRejected, reason)
		rec.QuarantineStatus = QuarantineStatusRejected
		rec.RejectionReason = reason
		return rec, err
	}

	if err := s.provider.Put(ctx, rec.StorageKey, bytes.NewReader(buf), int64(len(buf)), mimeType); err != nil {
		return nil, fmt.Errorf("failed to store clean file: %w", err)
	}

	_ = s.repo.UpdateQuarantineStatus(ctx, uploadID, QuarantineStatusClean, "")
	rec.QuarantineStatus = QuarantineStatusClean
	rec.MimeType = mimeType
	rec.FileSize = int64(len(buf))

	return rec, nil
}

func (s *UploadService) GetDownloadURL(ctx context.Context, requesterID string, uploadID string) (*DownloadResponse, error) {
	rec, err := s.repo.GetByID(ctx, uploadID)
	if err != nil {
		return nil, err
	}

	if rec.UserID != requesterID {
		return nil, ErrAccessDenied
	}

	if rec.QuarantineStatus != QuarantineStatusClean {
		return nil, ErrQuarantinedFile
	}

	expires := 15 * time.Minute
	downloadURL, err := s.provider.PresignedDownloadURL(ctx, rec.StorageKey, expires, rec.OriginalFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download url: %w", err)
	}

	return &DownloadResponse{
		DownloadURL: downloadURL,
		Filename:    rec.OriginalFilename,
		MimeType:    rec.MimeType,
		ExpiresAt:   time.Now().Add(expires),
	}, nil
}

func (s *UploadService) DeleteUpload(ctx context.Context, userID string, uploadID string) error {
	rec, err := s.repo.GetByID(ctx, uploadID)
	if err != nil {
		return err
	}

	if rec.UserID != userID {
		return ErrAccessDenied
	}

	if err := s.provider.Delete(ctx, rec.StorageKey); err != nil {
		return fmt.Errorf("failed to purge storage blob: %w", err)
	}

	return s.repo.Delete(ctx, uploadID)
}
