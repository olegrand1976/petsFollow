package invoicing

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	secretPrefixPlain = "plain:"
	secretPrefixEnc   = "enc:v1:"
)

// SealAPIKey stores a practice API key according to backend.
func SealAPIKey(backend, keyMaterial, apiKey string) (string, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return "", errors.New("empty api key")
	}
	switch backend {
	case "local_enc":
		if strings.TrimSpace(keyMaterial) == "" {
			return "", errors.New("BILLIT_SECRETS_KEY required for local_enc")
		}
		ct, err := encryptAESGCM(deriveKey(keyMaterial), []byte(apiKey))
		if err != nil {
			return "", err
		}
		return secretPrefixEnc + base64.RawStdEncoding.EncodeToString(ct), nil
	default: // plain_dev — local/mock only
		return secretPrefixPlain + apiKey, nil
	}
}

// OpenAPIKey recovers the practice API key from a sealed ref.
// Decrypt / unknown-ref failures wrap ErrSecrets so handlers can ask for reconnect.
func OpenAPIKey(backend, keyMaterial, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	switch {
	case strings.HasPrefix(ref, secretPrefixPlain):
		return strings.TrimPrefix(ref, secretPrefixPlain), nil
	case strings.HasPrefix(ref, secretPrefixEnc):
		raw, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(ref, secretPrefixEnc))
		if err != nil {
			return "", fmt.Errorf("%w: decode", ErrSecrets)
		}
		if strings.TrimSpace(keyMaterial) == "" {
			return "", fmt.Errorf("%w: BILLIT_SECRETS_KEY required", ErrSecrets)
		}
		pt, err := decryptAESGCM(deriveKey(keyMaterial), raw)
		if err != nil {
			return "", fmt.Errorf("%w: decrypt", ErrSecrets)
		}
		return string(pt), nil
	default:
		return "", fmt.Errorf("%w: unknown secret ref", ErrSecrets)
	}
}

func deriveKey(material string) []byte {
	sum := sha256.Sum256([]byte(material))
	return sum[:]
}

func encryptAESGCM(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptAESGCM(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ct := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, nil)
}

// NewConnectState returns a random opaque state token.
func NewConnectState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
