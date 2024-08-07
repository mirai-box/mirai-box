package service

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/chacha20poly1305"
)

// GeneratePublicID generates an encrypted ArtID from the RevisionID and UserID.
func GeneratePublicID(internalID, userID uuid.UUID, secretKey []byte) (string, error) {
	data := fmt.Sprintf("%s:%s", internalID.String(), userID.String())
	aead, err := chacha20poly1305.NewX(secretKey)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	encrypted := aead.Seal(nonce, nonce, []byte(data), nil)
	return base58.Encode(encrypted), nil
}

// DecodeArtID decrypts the ArtID to retrieve the RevisionID and UserID.
func DecodePublicID(publicID string, secretKey []byte) (string, string, error) {
	encrypted, err := base58.Decode(publicID)
	if err != nil {
		slog.Error("invalid ArtID format", "error", err, "publicID", publicID)
		return "", "", fmt.Errorf("invalid publicID format: %w", err)
	}

	aead, err := chacha20poly1305.NewX(secretKey)
	if err != nil {
		return "", "", err
	}

	if len(encrypted) < aead.NonceSize() {
		return "", "", fmt.Errorf("invalid publicID length")
	}

	nonce, ciphertext := encrypted[:aead.NonceSize()], encrypted[aead.NonceSize():]
	decrypted, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		slog.Error("can't decrypts and authenticates ciphertext", "error", err, "publicID", publicID)
		return "", "", err
	}

	data := strings.Split(string(decrypted), ":")
	if len(data) != 2 {
		slog.Error("can't get data from decrypted string", "error", err, "decrypted", decrypted)
		return "", "", fmt.Errorf("invalid publicID format")
	}

	return data[0], data[1], nil
}

// HashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
