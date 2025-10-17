package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

func SignResource(secret string, parts ...string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	payload := strings.Join(parts, ":")
	mac.Write([]byte(payload))
	sum := mac.Sum(nil)
	return []byte(base64.RawURLEncoding.EncodeToString(sum))
}
