package provider

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EncryptSecret encrypts plaintext using AES-256-GCM with authenticated additional data.
func EncryptSecret(masterKey []byte, plaintext []byte, additionalData []byte) (*EncryptedPayload, error) {
	if len(masterKey) != 32 {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM AEAD: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate random nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, additionalData)

	return &EncryptedPayload{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		KeyVersion: "v1",
	}, nil
}

// DecryptSecret decrypts an EncryptedPayload using AES-256-GCM.
func DecryptSecret(masterKey []byte, payload *EncryptedPayload, additionalData []byte) ([]byte, error) {
	if len(masterKey) != 32 {
		return nil, ErrInvalidKey
	}
	if payload == nil || len(payload.Ciphertext) == 0 || len(payload.Nonce) == 0 {
		return nil, ErrDecryptionFailed
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM AEAD: %w", err)
	}

	plaintext, err := gcm.Open(nil, payload.Nonce, payload.Ciphertext, additionalData)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// MaskSecret generates a safe, non-reversible visual preview of a secret.
func MaskSecret(secret string) string {
	if len(secret) <= 8 {
		return "********"
	}
	prefix := secret[:3]
	suffix := secret[len(secret)-4:]
	return fmt.Sprintf("%s****%s", prefix, suffix)
}

// CredentialVaultRepository defines storage persistence for encrypted credentials.
type CredentialVaultRepository interface {
	Save(ctx context.Context, cred *CredentialRecord) error
	GetByID(ctx context.Context, id string) (*CredentialRecord, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*CredentialRecord, error)
	Delete(ctx context.Context, id string) error
}

// MemoryCredentialVaultRepository provides thread-safe in-memory storage for credentials.
type MemoryCredentialVaultRepository struct {
	mu          sync.RWMutex
	credentials map[string]*CredentialRecord
}

func NewMemoryCredentialVaultRepository() *MemoryCredentialVaultRepository {
	return &MemoryCredentialVaultRepository{
		credentials: make(map[string]*CredentialRecord),
	}
}

func (r *MemoryCredentialVaultRepository) Save(ctx context.Context, cred *CredentialRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clone := *cred
	r.credentials[cred.ID] = &clone
	return nil
}

func (r *MemoryCredentialVaultRepository) GetByID(ctx context.Context, id string) (*CredentialRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cred, ok := r.credentials[id]
	if !ok {
		return nil, ErrCredentialNotFound
	}
	clone := *cred
	return &clone, nil
}

func (r *MemoryCredentialVaultRepository) ListByWorkspace(ctx context.Context, workspaceID string) ([]*CredentialRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*CredentialRecord
	for _, cred := range r.credentials {
		if cred.WorkspaceID == workspaceID {
			clone := *cred
			list = append(list, &clone)
		}
	}
	return list, nil
}

func (r *MemoryCredentialVaultRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.credentials[id]; !ok {
		return ErrCredentialNotFound
	}
	delete(r.credentials, id)
	return nil
}

// CredentialVaultService coordinates encryption, owner-isolation, and secret lifecycle (AT-016).
type CredentialVaultService struct {
	repo      CredentialVaultRepository
	masterKey []byte
}

func NewCredentialVaultService(repo CredentialVaultRepository, masterKey []byte) (*CredentialVaultService, error) {
	if len(masterKey) != 32 {
		return nil, ErrInvalidKey
	}
	return &CredentialVaultService{
		repo:      repo,
		masterKey: masterKey,
	}, nil
}

// StoreCredential encrypts and saves an account credential.
func (s *CredentialVaultService) StoreCredential(
	ctx context.Context,
	workspaceID, ownerID, providerID, accountID, rawSecret string,
	scopes []string,
	expiresAt *time.Time,
) (*CredentialMetadata, error) {
	if workspaceID == "" || ownerID == "" || providerID == "" || rawSecret == "" {
		return nil, fmt.Errorf("workspaceID, ownerID, providerID and rawSecret are required")
	}

	credID := uuid.New().String()
	aad := []byte(fmt.Sprintf("%s:%s:%s", workspaceID, ownerID, providerID))

	encrypted, err := EncryptSecret(s.masterKey, []byte(rawSecret), aad)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	record := &CredentialRecord{
		ID:                credID,
		WorkspaceID:       workspaceID,
		OwnerID:           ownerID,
		ProviderID:        providerID,
		AccountIdentifier: strings.TrimSpace(accountID),
		MaskedSecret:      MaskSecret(rawSecret),
		EncryptedSecret:   encrypted,
		Scopes:            scopes,
		Status:            "active",
		ExpiresAt:         expiresAt,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.Save(ctx, record); err != nil {
		return nil, err
	}

	meta := record.ToMetadata()
	return &meta, nil
}

// GetCredentialMetadata returns non-sensitive metadata for a credential (AT-016).
func (s *CredentialVaultService) GetCredentialMetadata(ctx context.Context, id, requesterID string) (*CredentialMetadata, error) {
	cred, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Owner isolation check: only owner or system can inspect
	if requesterID != "" && cred.OwnerID != requesterID && requesterID != "system" {
		return nil, ErrAccessDenied
	}

	meta := cred.ToMetadata()
	return &meta, nil
}

// ListWorkspaceCredentials returns redacted credentials for a workspace (AT-016).
func (s *CredentialVaultService) ListWorkspaceCredentials(ctx context.Context, workspaceID, requesterID string) ([]CredentialMetadata, error) {
	records, err := s.repo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var out []CredentialMetadata
	for _, cred := range records {
		// Only list credentials owned by requester or if requester is system/admin
		if requesterID == "" || cred.OwnerID == requesterID || requesterID == "system" {
			out = append(out, cred.ToMetadata())
		}
	}
	return out, nil
}

// GetDecryptedSecret retrieves and decrypts the secret for execution (owner or system only).
func (s *CredentialVaultService) GetDecryptedSecret(ctx context.Context, id, requesterID string) (string, []string, error) {
	cred, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", nil, err
	}

	if cred.Status != "active" {
		return "", nil, fmt.Errorf("credential %s is %s", id, cred.Status)
	}

	if cred.ExpiresAt != nil && cred.ExpiresAt.Before(time.Now().UTC()) {
		return "", nil, fmt.Errorf("credential %s is expired", id)
	}

	// Owner isolation check: only credential owner or authorized background execution engine can decrypt
	if requesterID != "" && cred.OwnerID != requesterID && requesterID != "system" {
		return "", nil, ErrAccessDenied
	}

	aad := []byte(fmt.Sprintf("%s:%s:%s", cred.WorkspaceID, cred.OwnerID, cred.ProviderID))
	plaintext, err := DecryptSecret(s.masterKey, cred.EncryptedSecret, aad)
	if err != nil {
		return "", nil, err
	}

	return string(plaintext), cred.Scopes, nil
}

// RevokeCredential changes status to revoked and removes the encrypted secret.
func (s *CredentialVaultService) RevokeCredential(ctx context.Context, id, requesterID string) error {
	cred, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if requesterID != "" && cred.OwnerID != requesterID && requesterID != "system" {
		return ErrAccessDenied
	}

	cred.Status = "revoked"
	cred.EncryptedSecret = nil
	cred.UpdatedAt = time.Now().UTC()

	return s.repo.Save(ctx, cred)
}
