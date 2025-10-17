package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	HeaderSignature = "X-Codex-Signature"
	HeaderDate      = "X-Codex-Date"
	HeaderNonce     = "X-Codex-Nonce"
)

func ComputeBodyHash(body []byte) string {
	sum := sha256.Sum256(body)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func ComputeSignature(secret string, tokenID string, method string, path string, query string, bodyHash string, date string, nonce string) string {
	data := strings.Join([]string{
		tokenID,
		strings.ToUpper(method),
		path,
		query,
		bodyHash,
		date,
		nonce,
	}, "\n")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func ValidateSignature(secret string, tokenID string, signature string, method string, path string, query string, body []byte, date string, nonce string) bool {
	bodyHash := ComputeBodyHash(body)
	expected := ComputeSignature(secret, tokenID, method, path, query, bodyHash, date, nonce)
	return hmac.Equal([]byte(signature), []byte(expected))
}

func ExtractSignatureHeaders(c *gin.Context) (date string, nonce string, signature string, err error) {
	date = c.GetHeader(HeaderDate)
	nonce = c.GetHeader(HeaderNonce)
	signature = c.GetHeader(HeaderSignature)

	if date == "" || nonce == "" || signature == "" {
		return "", "", "", fmt.Errorf("missing signature headers")
	}
	return date, nonce, signature, nil
}

func CanonicalPath(r *http.Request) (string, string) {
	path := r.URL.Path
	query := r.URL.RawQuery
	return path, query
}
