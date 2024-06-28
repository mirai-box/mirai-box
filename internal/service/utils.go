package service

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
)

const base62Chars = "OC4XiPTFk8Jx370lrWoLUEm6wvcye2tujdHsQqh5BzVDaSNYf1G9AgZKRnIbMp"

// encodeBase62 encodes a byte slice to a base62 string
func encodeBase62(data []byte) string {
	var result []byte
	for _, byteVal := range data {
		result = append(result, base62Chars[byteVal%62])
	}
	return string(result)
}

func getArtID(picID, revID string) string {
	combinedStr := picID + ":" + revID

	// Compute the SHA-256 hash
	hash := sha256.New()
	hash.Write([]byte(combinedStr))
	hashBytes := hash.Sum(nil)

	base62Hash := encodeBase62(hashBytes)

	return base62Hash
}

func getFilePath(filename, pictureID string, version int) string {
	extension := filepath.Ext(filename)
	filePath := filepath.Join(pictureID, fmt.Sprintf("%03d%s", version, extension))
	return filePath
}
