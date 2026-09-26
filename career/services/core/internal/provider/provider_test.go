package provider

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestVault_AES256GCM_EncryptionDecryption_AT016(t *testing.T) {
	masterKey := []byte("12345678901234567890123456789012") // 32 bytes
	plaintext := []byte("oauth_refresh_token_very_secret_12345")
	aad := []byte("ws-1:owner-1:linkedin")

	// 1. Successful encryption
	encrypted, err := EncryptSecret(masterKey, plaintext, aad)
	if err != nil {
		t.Fatalf("EncryptSecret failed: %v", err)
	}
	if len(encrypted.Ciphertext) == 0 || len(encrypted.Nonce) == 0 {
		t.Fatalf("Encrypted payload has empty ciphertext or nonce")
	}

	// 2. Successful decryption
	decrypted, err := DecryptSecret(masterKey, encrypted, aad)
	if err != nil {
		t.Fatalf("DecryptSecret failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Fatalf("Decrypted content mismatch: got %s, want %s", string(decrypted), string(plaintext))
	}

	// 3. Tampered ciphertext rejection (AEAD integrity)
	tamperedCipher := make([]byte, len(encrypted.Ciphertext))
	copy(tamperedCipher, encrypted.Ciphertext)
	tamperedCipher[0] ^= 0xFF // flip bits
	tamperedPayload := &EncryptedPayload{
		Ciphertext: tamperedCipher,
		Nonce:      encrypted.Nonce,
		KeyVersion: encrypted.KeyVersion,
	}
	_, err = DecryptSecret(masterKey, tamperedPayload, aad)
	if err != ErrDecryptionFailed {
		t.Fatalf("Expected ErrDecryptionFailed on tampered ciphertext, got %v", err)
	}

	// 4. Tampered AAD (Context binding)
	wrongAAD := []byte("ws-2:owner-2:linkedin")
	_, err = DecryptSecret(masterKey, encrypted, wrongAAD)
	if err != ErrDecryptionFailed {
		t.Fatalf("Expected ErrDecryptionFailed on wrong AAD, got %v", err)
	}

	// 5. Invalid key size rejection
	invalidKey := []byte("too-short-key")
	_, err = EncryptSecret(invalidKey, plaintext, aad)
	if err != ErrInvalidKey {
		t.Fatalf("Expected ErrInvalidKey, got %v", err)
	}
}

func TestVault_RedactionAndOwnerIsolation_AT016(t *testing.T) {
	ctx := context.Background()
	masterKey := []byte("12345678901234567890123456789012")
	repo := NewMemoryCredentialVaultRepository()
	service, err := NewCredentialVaultService(repo, masterKey)
	if err != nil {
		t.Fatalf("NewCredentialVaultService failed: %v", err)
	}

	secretToken := "secret_access_token_abcdef123456"
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	// 1. Store credential
	credMeta, err := service.StoreCredential(
		ctx,
		"ws-100",
		"owner-alice",
		"linkedin",
		"alice@example.com",
		secretToken,
		[]string{"profile", "w_member_social"},
		&expiresAt,
	)
	if err != nil {
		t.Fatalf("StoreCredential failed: %v", err)
	}

	// 2. Verify metadata redaction: secretToken must NOT appear in metadata
	if strings.Contains(credMeta.MaskedSecret, secretToken) {
		t.Fatalf("MaskedSecret leaked plaintext secret: %s", credMeta.MaskedSecret)
	}
	if !strings.HasPrefix(credMeta.MaskedSecret, "sec") || !strings.HasSuffix(credMeta.MaskedSecret, "3456") {
		t.Fatalf("MaskedSecret format unexpected: %s", credMeta.MaskedSecret)
	}

	// 3. Unauthorized user (Bob) is denied access
	_, err = service.GetCredentialMetadata(ctx, credMeta.ID, "user-bob")
	if err != ErrAccessDenied {
		t.Fatalf("Expected ErrAccessDenied for Bob, got %v", err)
	}
	_, _, err = service.GetDecryptedSecret(ctx, credMeta.ID, "user-bob")
	if err != ErrAccessDenied {
		t.Fatalf("Expected ErrAccessDenied when Bob tries to decrypt, got %v", err)
	}

	// 4. Authorized owner (Alice) and System can retrieve/decrypt
	decryptedSecret, scopes, err := service.GetDecryptedSecret(ctx, credMeta.ID, "owner-alice")
	if err != nil {
		t.Fatalf("Owner decryption failed: %v", err)
	}
	if decryptedSecret != secretToken {
		t.Fatalf("Decrypted secret mismatch: got %s, want %s", decryptedSecret, secretToken)
	}
	if len(scopes) != 2 {
		t.Fatalf("Expected 2 scopes, got %d", len(scopes))
	}

	// 5. Revocation
	err = service.RevokeCredential(ctx, credMeta.ID, "owner-alice")
	if err != nil {
		t.Fatalf("RevokeCredential failed: %v", err)
	}

	// Decryption after revocation must fail
	_, _, err = service.GetDecryptedSecret(ctx, credMeta.ID, "owner-alice")
	if err == nil {
		t.Fatalf("Expected error decrypting revoked credential, got nil")
	}
}

func TestCapabilityRegistry_GatingAndUnsupportedActions_AT010(t *testing.T) {
	registry := NewCapabilityRegistry()

	// 1. Supported action with granted scope -> success
	err := registry.AssertCapability("linkedin", "linkedin.post.create", []string{"openid", "w_member_social"})
	if err != nil {
		t.Fatalf("Expected capability to pass, got: %v", err)
	}

	// 2. Missing required scope -> fails closed with ErrCapabilityNotGranted (AT-010)
	err = registry.AssertCapability("linkedin", "linkedin.post.create", []string{"openid", "profile"})
	if err == nil || !strings.Contains(err.Error(), "requires scope 'w_member_social'") {
		t.Fatalf("Expected ErrCapabilityNotGranted for missing w_member_social, got: %v", err)
	}

	// 3. Unsupported action -> fails closed with ErrActionUnsupported (AT-010: no fake simulation)
	err = registry.AssertCapability("linkedin", "linkedin.direct_message.bulk_spam", []string{"openid", "w_member_social"})
	if err == nil || !strings.Contains(err.Error(), "is not supported by provider") {
		t.Fatalf("Expected ErrActionUnsupported for unsupported action, got: %v", err)
	}

	// 4. Unknown provider -> error
	err = registry.AssertCapability("unknown_platform", "any.action", nil)
	if err == nil {
		t.Fatalf("Expected error for unknown provider, got nil")
	}
}
