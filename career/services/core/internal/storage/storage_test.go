package storage

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStorage_FilenameSanitization(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"../../etc/passwd", "passwd"},
		{"..\\..\\boot.ini", "boot.ini"},
		{"normal_resume.pdf", "normal_resume.pdf"},
		{"bad/path\\name with space.docx", "name_with_space.docx"},
		{"../../../", "upload.bin"},
	}

	for _, c := range cases {
		got := SanitizeFilename(c.input)
		if got != c.expected {
			t.Errorf("SanitizeFilename(%q) = %q; want %q", c.input, got, c.expected)
		}
	}
}

func TestStorage_AT020_UploadBombAndSizeLimits(t *testing.T) {
	hugeDeclared := int64(100 * 1024 * 1024)
	_, err := ValidateFileHeader("huge.pdf", []byte("%PDF-1.4"), hugeDeclared)
	if err != ErrFileTooLarge {
		t.Fatalf("expected ErrFileTooLarge for declared huge size, got %v", err)
	}

	largeStream := bytes.NewReader(make([]byte, DefaultMaxUploadSize+100))
	_, _, _, err = InspectStream(largeStream)
	if err != ErrFileTooLarge {
		t.Fatalf("expected ErrFileTooLarge for large stream inspection, got %v", err)
	}
}

func TestStorage_QuarantineDetection(t *testing.T) {
	encryptedPDF := []byte("%PDF-1.5\n1 0 obj\n<< /Encrypt 2 0 R >>\nendobj\n")
	status, reason := ScanForProhibitedContent(encryptedPDF)
	if status != QuarantineStatusRejected {
		t.Fatalf("expected encrypted PDF to be rejected, got %s (reason: %s)", status, reason)
	}

	scriptPDF := []byte("%PDF-1.4\n<< /JavaScript (alert(1)) >>\n")
	status, reason = ScanForProhibitedContent(scriptPDF)
	if status != QuarantineStatusRejected {
		t.Fatalf("expected script PDF to be rejected, got %s", status)
	}

	macroDOCX := append([]byte("PK\x03\x04"), []byte("...vbaProject.bin...")...)
	status, reason = ScanForProhibitedContent(macroDOCX)
	if status != QuarantineStatusRejected {
		t.Fatalf("expected macro DOCX to be rejected, got %s", status)
	}

	cleanPDF := []byte("%PDF-1.4\nClean Resume Content\n%%EOF")
	status, _ = ScanForProhibitedContent(cleanPDF)
	if status != QuarantineStatusClean {
		t.Fatalf("expected clean PDF to pass, got %s", status)
	}
}

func TestStorage_LocalStorageProvider_PutGetDeleteToken(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	secretKey := []byte("super-secret-signing-key-32-byte!")
	provider, err := NewLocalStorageProvider(tempDir, secretKey, "http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	key := "resumes/user1/test_resume.pdf"
	content := []byte("%PDF-1.4 Sample Resume")

	err = provider.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "application/pdf")
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	exists, err := provider.Exists(ctx, key)
	if err != nil || !exists {
		t.Fatalf("expected Exists=true, got %v (err: %v)", exists, err)
	}

	rc, size, mime, err := provider.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if size != int64(len(content)) || mime != "application/pdf" {
		t.Errorf("Get returned size=%d, mime=%s", size, mime)
	}
	_ = rc.Close()

	token := provider.GenerateSignedToken("download", key, 5*time.Minute)
	if err := provider.ValidateSignedToken(token, "download", key); err != nil {
		t.Errorf("valid token validation failed: %v", err)
	}

	if err := provider.ValidateSignedToken(token, "upload", key); err == nil {
		t.Error("expected action mismatch error, got nil")
	}

	expiredToken := provider.GenerateSignedToken("download", key, -1*time.Minute)
	if err := provider.ValidateSignedToken(expiredToken, "download", key); err == nil {
		t.Error("expected expired token error, got nil")
	}

	if err := provider.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	exists, _ = provider.Exists(ctx, key)
	if exists {
		t.Error("expected file to be deleted")
	}
}

func TestStorage_UploadServiceLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "svc-storage-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	provider, err := NewLocalStorageProvider(tempDir, []byte("secret-token-key-123456789012"), "http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}

	repo := NewMemoryUploadRepository()
	service := NewUploadService(repo, provider)

	ctx := context.Background()
	userID := uuid.New().String()
	otherUserID := uuid.New().String()

	upResp, err := service.RequestUpload(ctx, userID, UploadRequest{
		Filename:    "My_Resume.pdf",
		ContentType: "application/pdf",
		Size:        1024,
	})
	if err != nil {
		t.Fatalf("RequestUpload failed: %v", err)
	}

	_, err = service.GetDownloadURL(ctx, userID, upResp.UploadID)
	if err != ErrQuarantinedFile {
		t.Fatalf("expected ErrQuarantinedFile before processing, got %v", err)
	}

	cleanData := []byte("%PDF-1.4\nJohn Doe Resume Text\n%%EOF")
	record, err := service.ProcessAndCommitUpload(ctx, upResp.UploadID, bytes.NewReader(cleanData))
	if err != nil {
		t.Fatalf("ProcessAndCommitUpload failed: %v", err)
	}
	if record.QuarantineStatus != QuarantineStatusClean {
		t.Fatalf("expected status clean, got %s", record.QuarantineStatus)
	}

	dlResp, err := service.GetDownloadURL(ctx, userID, upResp.UploadID)
	if err != nil {
		t.Fatalf("GetDownloadURL failed for owner: %v", err)
	}
	if !strings.Contains(dlResp.DownloadURL, "/api/v1/storage/download?token=") {
		t.Fatalf("unexpected download URL: %s", dlResp.DownloadURL)
	}

	_, err = service.GetDownloadURL(ctx, otherUserID, upResp.UploadID)
	if err != ErrAccessDenied {
		t.Fatalf("expected ErrAccessDenied for other user, got %v", err)
	}

	err = service.DeleteUpload(ctx, userID, upResp.UploadID)
	if err != nil {
		t.Fatalf("DeleteUpload failed: %v", err)
	}

	_, err = service.GetDownloadURL(ctx, userID, upResp.UploadID)
	if err != ErrUploadNotFound {
		t.Fatalf("expected ErrUploadNotFound after deletion, got %v", err)
	}
}
