package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
	"time"
)

// Nonce generates a base64-encoded nonce for Xiaomi API requests.
// Format: 8 random bytes + minutes-since-epoch bytes, base64-encoded.
func Nonce() (string, error) {
	// Generate 8 random bytes
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Convert to signed int64 range: [-(2^63), 2^63-1]
	n := new(big.Int).SetBytes(randomBytes)
	maxVal := new(big.Int).Exp(big.NewInt(2), big.NewInt(63), nil)
	n.Sub(n, maxVal)
	randomSigned := n.Bytes()

	// Minutes since epoch
	millis := time.Now().UnixMilli()
	part2 := millis / 60000

	// Encode part2 as big-endian bytes (variable length based on bit length)
	part2Bytes := big.NewInt(part2).Bytes()

	// Concatenate
	combined := make([]byte, 0, len(randomSigned)+len(part2Bytes))
	combined = append(combined, randomSigned...)
	combined = append(combined, part2Bytes...)

	return base64.StdEncoding.EncodeToString(combined), nil
}

// SignedNonce computes the signed nonce from ssecurity and nonce.
// SHA-256(base64_decode(ssecurity) + base64_decode(nonce)), base64-encoded.
func SignedNonce(ssecurity, nonce string) (string, error) {
	ssecurityBytes, err := base64.StdEncoding.DecodeString(ssecurity)
	if err != nil {
		return "", err
	}
	nonceBytes, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(ssecurityBytes)
	h.Write(nonceBytes)

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
