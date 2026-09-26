package provider

import (
	"errors"
	"time"
)

var (
	ErrCapabilityNotGranted = errors.New("capability not granted by credential scopes")
	ErrActionUnsupported    = errors.New("action is unsupported by provider manifest")
	ErrCredentialNotFound   = errors.New("credential not found")
	ErrDecryptionFailed     = errors.New("decryption failed: invalid key or tampered ciphertext")
	ErrAccessDenied         = errors.New("access denied to credential")
	ErrInvalidKey           = errors.New("invalid master encryption key: must be 32 bytes for AES-256")
	ErrProviderNotFound     = errors.New("provider manifest not found")
)

// ActionCapability defines an action supported by a provider and the required OAuth scope.
type ActionCapability struct {
	Action        string `json:"action"`
	RequiredScope string `json:"required_scope"`
	Description   string `json:"description"`
}

// ProviderManifest specifies permitted actions, scopes, and upstream security review metadata (REQ-021, REQ-024).
type ProviderManifest struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Channel         string             `json:"channel"` // e.g. "linkedin", "google_sheets", "job_board"
	SupportedScopes []string           `json:"supported_scopes"`
	Capabilities    []ActionCapability `json:"capabilities"`
	ReviewDate      string             `json:"review_date"` // YYYY-MM-DD
	Enabled         bool               `json:"enabled"`
	DocumentationURL string            `json:"documentation_url"`
}

// EncryptedPayload encapsulates AES-256-GCM encrypted data and initialization nonce.
type EncryptedPayload struct {
	Ciphertext []byte `json:"ciphertext"`
	Nonce      []byte `json:"nonce"`
	KeyVersion string `json:"key_version"`
}

// CredentialRecord represents an encrypted, owner-isolated credential stored in the vault (AT-016).
type CredentialRecord struct {
	ID                string            `json:"id"`
	WorkspaceID       string            `json:"workspace_id"`
	OwnerID           string            `json:"owner_id"`
	ProviderID        string            `json:"provider_id"`
	AccountIdentifier string            `json:"account_identifier"` // e.g., email or handle
	MaskedSecret      string            `json:"masked_secret"`      // e.g., "tok_****_abc"
	EncryptedSecret   *EncryptedPayload `json:"-"`                  // Never serialized to JSON
	Scopes            []string          `json:"scopes"`
	Status            string            `json:"status"` // "active", "expired", "revoked"
	ExpiresAt         *time.Time        `json:"expires_at,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// CredentialMetadata is the redacted user-facing representation of a credential (AT-016).
type CredentialMetadata struct {
	ID                string     `json:"id"`
	WorkspaceID       string     `json:"workspace_id"`
	OwnerID           string     `json:"owner_id"`
	ProviderID        string     `json:"provider_id"`
	AccountIdentifier string     `json:"account_identifier"`
	MaskedSecret      string     `json:"masked_secret"`
	Scopes            []string   `json:"scopes"`
	Status            string     `json:"status"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ToMetadata converts an internal CredentialRecord to a safe, redacted CredentialMetadata.
func (c *CredentialRecord) ToMetadata() CredentialMetadata {
	return CredentialMetadata{
		ID:                c.ID,
		WorkspaceID:       c.WorkspaceID,
		OwnerID:           c.OwnerID,
		ProviderID:        c.ProviderID,
		AccountIdentifier: c.AccountIdentifier,
		MaskedSecret:      c.MaskedSecret,
		Scopes:            c.Scopes,
		Status:            c.Status,
		ExpiresAt:         c.ExpiresAt,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}
