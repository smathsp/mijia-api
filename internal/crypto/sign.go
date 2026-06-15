package crypto

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
)

// EncSignature generates the encryption signature for API requests.
// Builds: METHOD&uri&k=v&...&signed_nonce, then SHA-1 hashes and base64-encodes.
func EncSignature(uri, method, signedNonce string, params map[string]string) string {
	signatureParams := []string{
		strings.ToUpper(method),
		uri,
	}

	// Add params in sorted key order for consistency
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		signatureParams = append(signatureParams, fmt.Sprintf("%s=%s", k, params[k]))
	}

	signatureParams = append(signatureParams, signedNonce)
	signatureString := strings.Join(signatureParams, "&")

	h := sha1.New()
	h.Write([]byte(signatureString))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// GenerateEncParams generates encrypted parameters for API requests.
// Steps:
// 1. Add rc4_hash__ = EncSignature(plaintext params)
// 2. RC4-encrypt all param values (including rc4_hash__)
// 3. Add signature = EncSignature(encrypted params)
// 4. Add ssecurity and _nonce
func GenerateEncParams(uri, method, signedNonce, nonce string, params map[string]string, ssecurity string) map[string]string {
	// Step 1: Compute signature of plaintext params
	params["rc4_hash__"] = EncSignature(uri, method, signedNonce, params)

	// Step 2: RC4-encrypt all values
	for k, v := range params {
		encrypted, err := EncryptRC4(signedNonce, v)
		if err != nil {
			// Fallback: keep original value (should not happen in practice)
			continue
		}
		params[k] = encrypted
	}

	// Step 3: Add signature of encrypted params + ssecurity + nonce
	params["signature"] = EncSignature(uri, method, signedNonce, params)
	params["ssecurity"] = ssecurity
	params["_nonce"] = nonce

	return params
}
