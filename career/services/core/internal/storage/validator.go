package storage

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	DefaultMaxUploadSize = 10 * 1024 * 1024 // 10MB
)

var allowedMIMETypes = map[string]bool{
	"application/pdf": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/msword": true,
	"text/plain":         true,
	"image/png":           true,
	"image/jpeg":          true,
	"image/webp":          true,
}

var safeFilenameRegex = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func SanitizeFilename(filename string) string {
	normalized := strings.ReplaceAll(filename, "\\", "/")
	parts := strings.Split(normalized, "/")
	base := ""
	for i := len(parts) - 1; i >= 0; i-- {
		p := strings.TrimSpace(parts[i])
		if p != "" && p != "." && p != ".." {
			base = p
			break
		}
	}
	if base == "" {
		return "upload.bin"
	}
	cleaned := safeFilenameRegex.ReplaceAllString(base, "_")
	cleaned = strings.Trim(cleaned, "_")
	if cleaned == "" || cleaned == "." {
		return "upload.bin"
	}
	return cleaned
}

func ValidateFileHeader(filename string, headerBytes []byte, declaredSize int64) (string, error) {
	if declaredSize > DefaultMaxUploadSize {
		return "", ErrFileTooLarge
	}

	detectedMIME := http.DetectContentType(headerBytes)
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		if !bytes.HasPrefix(headerBytes, []byte("%PDF-")) {
			return "", ErrInvalidMimeType
		}
		return "application/pdf", nil
	case ".docx":
		if !bytes.HasPrefix(headerBytes, []byte("PK\x03\x04")) {
			return "", ErrInvalidMimeType
		}
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document", nil
	case ".png":
		if !bytes.HasPrefix(headerBytes, []byte("\x89PNG\r\n\x1a\n")) {
			return "", ErrInvalidMimeType
		}
		return "image/png", nil
	case ".jpg", ".jpeg":
		if !bytes.HasPrefix(headerBytes, []byte("\xff\xd8\xff")) {
			return "", ErrInvalidMimeType
		}
		return "image/jpeg", nil
	case ".txt":
		return "text/plain", nil
	default:
		baseMIME := strings.Split(detectedMIME, ";")[0]
		if !allowedMIMETypes[baseMIME] {
			return "", ErrInvalidMimeType
		}
		return baseMIME, nil
	}
}

func ScanForProhibitedContent(data []byte) (QuarantineStatus, string) {
	if bytes.HasPrefix(data, []byte("%PDF-")) {
		if bytes.Contains(data, []byte("/Encrypt")) {
			return QuarantineStatusRejected, "encrypted or password-protected PDFs are prohibited"
		}
		if bytes.Contains(data, []byte("/JavaScript")) || bytes.Contains(data, []byte("/Launch")) {
			return QuarantineStatusRejected, "PDF contains active scripts or launch actions"
		}
		return QuarantineStatusClean, ""
	}

	if bytes.HasPrefix(data, []byte("PK\x03\x04")) {
		if bytes.Contains(data, []byte("vbaProject.bin")) || bytes.Contains(data, []byte("vbaData.xml")) {
			return QuarantineStatusRejected, "documents containing executable macros are prohibited"
		}
		if bytes.Contains(data, []byte("EncryptedPackage")) {
			return QuarantineStatusRejected, "encrypted or password-protected Office documents are prohibited"
		}
		return QuarantineStatusClean, ""
	}

	return QuarantineStatusClean, ""
}

func InspectStream(r io.Reader) ([]byte, QuarantineStatus, string, error) {
	lr := io.LimitReader(r, DefaultMaxUploadSize+1)
	buf, err := io.ReadAll(lr)
	if err != nil {
		return nil, QuarantineStatusRejected, "read error during quarantine inspection", err
	}
	if int64(len(buf)) > DefaultMaxUploadSize {
		return nil, QuarantineStatusRejected, "file size exceeds maximum allowed limit", ErrFileTooLarge
	}

	status, reason := ScanForProhibitedContent(buf)
	return buf, status, reason, nil
}
