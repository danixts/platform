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

var (
	ErrInvalidKey        = errors.New("platform/crypto/aesgcm: invalid secret; must be base64 of exactly 32 bytes")
	ErrCipherTooShort    = errors.New("platform/crypto/aesgcm: ciphertext too short")
	ErrPlaceholderSecret = errors.New("platform/crypto/aesgcm: placeholder secret detected")
)

const PlaceholderMarker = "<<REPLACE_WITH_ENCRYPTED>>"

type Cipher struct {
	aead cipher.AEAD
}

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

func (c *Cipher) EncryptString(s string) (string, error) {
	return c.Encrypt([]byte(s))
}

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

func (c *Cipher) DecryptString(b64 string) (string, error) {
	b, err := c.Decrypt(b64)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func IsPlaceholder(b64 string) bool { return b64 == PlaceholderMarker }
