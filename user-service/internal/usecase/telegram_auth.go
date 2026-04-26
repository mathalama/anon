package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// VerifyTelegramHash validates data sent by Telegram Login Widget.
// See: https://core.telegram.org/widgets/login#checking-authorization
func VerifyTelegramHash(data map[string]string, botToken string) error {
	hash := data["hash"]
	if hash == "" {
		return fmt.Errorf("hash is missing")
	}

	// Create data-check-string
	var keys []string
	for k := range data {
		if k != "hash" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var dataCheckArr []string
	for _, k := range keys {
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", k, data[k]))
	}
	dataCheckString := strings.Join(dataCheckArr, "\n")

	// Calculate secret key: SHA256(botToken)
	sha := sha256.New()
	sha.Write([]byte(botToken))
	secretKey := sha.Sum(nil)

	// Calculate HMAC-SHA256 of data-check-string with secretKey
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	if expectedHash != hash {
		return fmt.Errorf("invalid hash: expected %s, got %s", expectedHash, hash)
	}

	return nil
}
