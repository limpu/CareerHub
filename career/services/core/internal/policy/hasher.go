package policy

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// CanonicalizeJSON parses raw JSON and formats it deterministically with sorted keys.
func CanonicalizeJSON(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return []byte("{}"), nil
	}

	var parsed interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("invalid json payload for canonicalization: %w", err)
	}

	canonical, err := json.Marshal(parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to re-serialize canonical json: %w", err)
	}

	return canonical, nil
}

// ComputePayloadHash computes the SHA-256 hex digest of canonicalized JSON (AT-007).
func ComputePayloadHash(payload json.RawMessage) (string, error) {
	canonical, err := CanonicalizeJSON(payload)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(canonical)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// ComputeDocumentHash computes the SHA-256 hex digest of a document blob (e.g. resume PDF).
func ComputeDocumentHash(docBytes []byte) string {
	if len(docBytes) == 0 {
		return ""
	}
	h := sha256.New()
	h.Write(docBytes)
	return fmt.Sprintf("%x", h.Sum(nil))
}
