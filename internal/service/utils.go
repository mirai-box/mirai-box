package service

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/chacha20poly1305"
)

// GenerateArtID generates an encrypted ArtID from the RevisionID and UserID.
func GenerateArtID(revisionID, userID uuid.UUID, secretKey []byte) (string, error) {
	slog.Debug("GenerateArtID: secretKey", "len", len(secretKey))

	data := fmt.Sprintf("%s:%s", revisionID.String(), userID.String())
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
func DecodeArtID(artID string, secretKey []byte) (string, string, error) {
	slog.Debug("DecodeArtID: secretKey", "len", len(secretKey))

	encrypted, err := base58.Decode(artID)
	if err != nil {
		slog.Error("invalid ArtID format", "error", err, "artID", artID)
		return "", "", fmt.Errorf("invalid ArtID format: %w", err)
	}

	aead, err := chacha20poly1305.NewX(secretKey)
	if err != nil {
		return "", "", err
	}

	if len(encrypted) < aead.NonceSize() {
		return "", "", fmt.Errorf("invalid ArtID length")
	}

	nonce, ciphertext := encrypted[:aead.NonceSize()], encrypted[aead.NonceSize():]
	decrypted, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		slog.Error("can't decrypts and authenticates ciphertext", "error", err, "artID", artID)
		return "", "", err
	}

	data := strings.Split(string(decrypted), ":")
	if len(data) != 2 {
		slog.Error("can't get data from decrypted string", "error", err, "decrypted", decrypted)
		return "", "", fmt.Errorf("invalid ArtID format")
	}

	return data[0], data[1], nil
}
