package save

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// hmacKey is embedded in the binary to sign save files.
// This deters casual tampering; it is not a cryptographic secret — a
// determined user could extract it from the binary. Server-side key
// management would be required for stronger guarantees (e.g. leaderboards).
const hmacKey = "clicker-save-hmac-v1-5f3a8c2d9e1b7f4a6c0e2d8b"

// sign computes HMAC-SHA256 over data and returns the result as a lowercase hex string.
func sign(data []byte) string {
	mac := hmac.New(sha256.New, []byte(hmacKey))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// verify reports whether sig is a valid HMAC-SHA256 signature of data.
// Uses constant-time comparison to prevent timing attacks.
func verify(data []byte, sig string) bool {
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(hmacKey))
	mac.Write(data)
	expected := mac.Sum(nil)
	return hmac.Equal(expected, sigBytes)
}
