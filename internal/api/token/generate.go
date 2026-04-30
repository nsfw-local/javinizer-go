package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const TokenPrefix = "jv_"

func GenerateToken() (fullToken string, tokenPrefix string, err error) {
	bytes := make([]byte, 16)
	if _, err = rand.Read(bytes); err != nil {
		return "", "", err
	}
	hexStr := hex.EncodeToString(bytes)
	fullToken = TokenPrefix + hexStr
	tokenPrefix = hexStr[:8]
	return fullToken, tokenPrefix, nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
