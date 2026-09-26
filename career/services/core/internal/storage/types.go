package storage

import (
	"errors"
	"time"
)

type QuarantineStatus string

const (
	QuarantineStatusQuarantined QuarantineStatus = "quarantine"
	QuarantineStatusScanning    QuarantineStatus = "scanning"
	QuarantineStatusClean       QuarantineStatus = "clean"
	QuarantineStatusRejected    QuarantineStatus = "rejected"
)

var (
	ErrFileTooLarge            = errors.New("file exceeds maximum allowed size")
	ErrInvalidMimeType         = errors.New("unsupported or mismatched file MIME type")
	ErrEncryptedOrMacroContent = errors.New("encrypted or macro-enabled documents are prohibited")
	ErrInvalidStorageToken     = errors.New("invalid, expired or forged storage token")
	ErrQuarantinedFile         = errors.New("file is in quarantine and cannot be downloaded")
	ErrUploadNotFound          = errors.New("upload record not found")
	ErrAccessDenied            = errors.New("access to storage resource denied")
)

type UploadRecord struct {
	ID               string           `json:"id"`
	UserID           string           `json:"user_id"`
	OriginalFilename string           `json:"original_filename"`
	StorageKey       string           `json:"storage_key"`
	MimeType         string           `json:"mime_type"`
	FileSize         int64            `json:"file_size"`
	QuarantineStatus QuarantineStatus `json:"quarantine_status"`
	RejectionReason  string           `json:"rejection_reason,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

type UploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type UploadResponse struct {
	UploadID   string    `json:"upload_id"`
	StorageKey string    `json:"storage_key"`
	UploadURL  string    `json:"upload_url"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type DownloadResponse struct {
	DownloadURL string    `json:"download_url"`
	Filename    string    `json:"filename"`
	MimeType    string    `json:"mime_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}