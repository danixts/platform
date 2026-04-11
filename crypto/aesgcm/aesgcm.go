// Package aesgcm provides an AES-256-GCM AEAD cipher keyed from a
// base64-encoded 32-byte secret (APP_SECRET). It is the standard
// primitive used by XMart Cloud services to encrypt sensitive columns
// (suffix "_enc") at rest.
package aesgcm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Error values returned by the package.
var (
	ErrInvalidKey        = errors.New("xmart-platform/crypto/aesgcm: invalid secret; must be base64 of exactly 32 bytes")
	ErrCipherTooShort    = errors.New("xmart-platform/crypto/aesgcm: ciphertext too short")
	ErrPlaceholderSecret = errors.New("xmart-platform/crypto/aesgcm: placeholder secret detected")
)

// PlaceholderMarker is the sentinel string stored in seed data when a
// ciphertext will be generated at first boot. Decrypt returns
// ErrPlaceholderSecret if it encounters this marker so callers can
// distinguish "needs encryption" from "real cipher error".
const PlaceholderMarker = "<<REPLACE_WITH_ENCRYPTED>>"

// Cipher is the ready-to-use AES-GCM primitive.
type Cipher struct {
	aead cipher.AEAD
}

// New builds a Cipher from a base64-encoded 32-byte key. Invalid input
// returns ErrInvalidKey without leaking details to the caller's logs.
func New(appSecretB64 string) (*Cipher, error) {
	key, err := base64.StdEncoding.DecodeString(appSecretB64)
	if err != nil || len(key) != 32 {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher.NewGCM: %w", err)
	}
	return &Cipher{aead: aead}, nil
}

// Encrypt returns a base64-raw-url-encoded ciphertext that embeds the
// random nonce as a prefix. The output is safe to store in URLs or
// JSON fields.
func (c *Cipher) Encrypt(plaintext []byte) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := c.aead.Seal(nil, nonce, plaintext, nil)
	buf := make([]byte, 0, len(nonce)+len(ct))
	buf = append(buf, nonce...)
	buf = append(buf, ct...)
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// EncryptString is Encrypt for convenience.
func (c *Cipher) EncryptString(s string) (string, error) {
	return c.Encrypt([]byte(s))
}

// Decrypt reverses Encrypt. It returns ErrPlaceholderSecret when the
// input is the sentinel PlaceholderMarker.
func (c *Cipher) Decrypt(b64 string) ([]byte, error) {
	if b64 == PlaceholderMarker {
		return nil, ErrPlaceholderSecret
	}
	raw, err := base64.RawURLEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	ns := c.aead.NonceSize()
	if len(raw) < ns+c.aead.Overhead() {
		return nil, ErrCipherTooShort
	}
	nonce, ct := raw[:ns], raw[ns:]
	return c.aead.Open(nil, nonce, ct, nil)
}

// DecryptString is Decrypt for convenience.
func (c *Cipher) DecryptString(b64 string) (string, error) {
	b, err := c.Decrypt(b64)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// IsPlaceholder reports whether a stored value is still the seed
// placeholder.
func IsPlaceholder(b64 string) bool { return b64 == PlaceholderMarker }
