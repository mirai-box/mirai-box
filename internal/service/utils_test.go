package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateArtID(t *testing.T) {
	secretKey := make([]byte, 32) // 32 bytes for XChaCha20-Poly1305
	for i := range secretKey {
		secretKey[i] = byte(i)
	}

	tests := []struct {
		name        string
		revisionID  uuid.UUID
		userID      uuid.UUID
		secretKey   []byte
		expectError bool
	}{
		{
			name:        "Valid input",
			revisionID:  uuid.New(),
			userID:      uuid.New(),
			secretKey:   secretKey,
			expectError: false,
		},
		{
			name:        "Invalid secret key length",
			revisionID:  uuid.New(),
			userID:      uuid.New(),
			secretKey:   []byte("too short"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artID, err := GeneratePublicID(tt.revisionID, tt.userID, tt.secretKey)
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, artID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, artID)
			}
		})
	}
}

func TestDecodeArtID(t *testing.T) {
	secretKey := make([]byte, 32) // 32 bytes for XChaCha20-Poly1305
	for i := range secretKey {
		secretKey[i] = byte(i)
	}

	revisionID := uuid.New()
	userID := uuid.New()

	validArtID, err := GeneratePublicID(revisionID, userID, secretKey)
	require.NoError(t, err)

	tests := []struct {
		name        string
		artID       string
		secretKey   []byte
		expectError bool
	}{
		{
			name:        "Valid input",
			artID:       validArtID,
			secretKey:   secretKey,
			expectError: false,
		},
		{
			name:        "Invalid ArtID",
			artID:       "invalid-art-id",
			secretKey:   secretKey,
			expectError: true,
		},
		{
			name:        "Invalid secret key",
			artID:       validArtID,
			secretKey:   []byte("wrong key"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decodedRevisionID, decodedUserID, err := DecodePublicID(tt.artID, tt.secretKey)
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, decodedRevisionID)
				assert.Empty(t, decodedUserID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, revisionID.String(), decodedRevisionID)
				assert.Equal(t, userID.String(), decodedUserID)
			}
		})
	}
}

func TestEndToEnd(t *testing.T) {
	secretKey := make([]byte, 32)
	for i := range secretKey {
		secretKey[i] = byte(i)
	}

	revisionID := uuid.New()
	userID := uuid.New()

	// Generate ArtID
	artID, err := GeneratePublicID(revisionID, userID, secretKey)
	require.NoError(t, err)
	assert.NotEmpty(t, artID)

	// Decode ArtID
	decodedRevisionID, decodedUserID, err := DecodePublicID(artID, secretKey)
	require.NoError(t, err)
	assert.Equal(t, revisionID.String(), decodedRevisionID)
	assert.Equal(t, userID.String(), decodedUserID)
}
