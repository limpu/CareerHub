package identity

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid or malformed token")
	ErrTokenExpired = errors.New("token has expired")
)

type JWTClaims struct {
	UserID       string       `json:"sub"`
	Email        string       `json:"email"`
	PlatformRole PlatformRole `json:"role"`
	SessionID    string       `json:"sid"`
	IssuedAt     int64        `json:"iat"`
	ExpiresAt    int64        `json:"exp"`
}

// GenerateSessionToken produces a cryptographically secure random session string.
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashSessionToken computes the SHA-256 hash of a session token for storage.
func HashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// SignJWT creates an HMAC-SHA256 signed JWT string for the given claims.
func SignJWT(claims JWTClaims, secret string) (string, error) {
	headerJSON, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	unsignedToken := fmt.Sprintf("%s.%s", headerB64, payloadB64)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsignedToken))
	sigB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%s", unsignedToken, sigB64), nil
}

// VerifyJWT parses and verifies the HMAC-SHA256 signature and expiration of a JWT string.
func VerifyJWT(tokenString, secret string) (*JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	unsignedToken := fmt.Sprintf("%s.%s", parts[0], parts[1])
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsignedToken))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(expectedSig, actualSig) {
		return nil, ErrInvalidToken
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().UTC().Unix() > claims.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}