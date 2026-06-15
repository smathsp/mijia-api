package crypto

import (
	"encoding/base64"
	"testing"
)

func TestNonce(t *testing.T) {
	nonce, err := Nonce()
	if err != nil {
		t.Fatalf("Nonce() error = %v", err)
	}

	// Verify it's valid base64
	decoded, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		t.Fatalf("Nonce() returned invalid base64: %v", err)
	}

	// Should be at least 8 bytes (random) + some time bytes
	if len(decoded) < 9 {
		t.Errorf("Nonce() too short: got %d bytes, want >= 9", len(decoded))
	}

	// Two nonces should be different (with extremely high probability)
	nonce2, _ := Nonce()
	if nonce == nonce2 {
		t.Error("Nonce() returned same value twice")
	}
}

func TestSignedNonce(t *testing.T) {
	ssecurity := "dGVzdHNlY3VyaXR5MTIz" // base64("testsecurity123")
	nonce := "dGVzdG5vbmNlNDU2"         // base64("testnonce456")

	signed, err := SignedNonce(ssecurity, nonce)
	if err != nil {
		t.Fatalf("SignedNonce() error = %v", err)
	}

	// Verify it's valid base64
	decoded, err := base64.StdEncoding.DecodeString(signed)
	if err != nil {
		t.Fatalf("SignedNonce() returned invalid base64: %v", err)
	}

	// SHA-256 produces 32 bytes
	if len(decoded) != 32 {
		t.Errorf("SignedNonce() length = %d, want 32", len(decoded))
	}

	// Same inputs should produce same output
	signed2, _ := SignedNonce(ssecurity, nonce)
	if signed != signed2 {
		t.Error("SignedNonce() not deterministic")
	}

	// Different inputs should produce different output
	signed3, _ := SignedNonce(ssecurity, "differentnonce")
	if signed == signed3 {
		t.Error("SignedNonce() same output for different inputs")
	}
}

func TestEncryptDecryptRC4(t *testing.T) {
	// Use a known key (base64 encoded)
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef"))

	plaintext := "hello world"
	encrypted, err := EncryptRC4(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptRC4() error = %v", err)
	}

	// Encrypted should be different from plaintext
	if encrypted == plaintext {
		t.Error("EncryptRC4() returned plaintext")
	}

	// Decrypt should return original
	decrypted, err := DecryptRC4(key, encrypted)
	if err != nil {
		t.Fatalf("DecryptRC4() error = %v", err)
	}

	if string(decrypted) != plaintext {
		t.Errorf("DecryptRC4() = %q, want %q", string(decrypted), plaintext)
	}
}

func TestEncryptRC4Deterministic(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef"))
	plaintext := "test data"

	encrypted1, _ := EncryptRC4(key, plaintext)
	encrypted2, _ := EncryptRC4(key, plaintext)

	// Same key + plaintext should produce same ciphertext
	if encrypted1 != encrypted2 {
		t.Error("EncryptRC4() not deterministic")
	}
}

func TestEncSignature(t *testing.T) {
	uri := "/miotspec/prop/get"
	method := "POST"
	signedNonce := "dGVzdHNpZ25lZG5vbmNl"
	params := map[string]string{
		"data":     `{"did":"12345"}`,
		"_nonce":   "dGVzdG5vbmNl",
	}

	sig1 := EncSignature(uri, method, signedNonce, params)
	sig2 := EncSignature(uri, method, signedNonce, params)

	// Same inputs should produce same signature
	if sig1 != sig2 {
		t.Error("EncSignature() not deterministic")
	}

	// Signature should be valid base64
	decoded, err := base64.StdEncoding.DecodeString(sig1)
	if err != nil {
		t.Fatalf("EncSignature() returned invalid base64: %v", err)
	}

	// SHA-1 produces 20 bytes
	if len(decoded) != 20 {
		t.Errorf("EncSignature() length = %d, want 20", len(decoded))
	}
}

func TestGenerateEncParams(t *testing.T) {
	uri := "/miotspec/prop/get"
	method := "POST"
	signedNonce := "dGVzdHNpZ25lZG5vbmNl"
	nonce := "dGVzdG5vbmNl"
	ssecurity := "dGVzdHNlY3VyaXR5"

	params := map[string]string{
		"data": `{"did":"12345"}`,
	}

	result := GenerateEncParams(uri, method, signedNonce, nonce, params, ssecurity)

	// Should have rc4_hash__, signature, ssecurity, _nonce
	if _, ok := result["rc4_hash__"]; !ok {
		t.Error("GenerateEncParams() missing rc4_hash__")
	}
	if _, ok := result["signature"]; !ok {
		t.Error("GenerateEncParams() missing signature")
	}
	if result["ssecurity"] != ssecurity {
		t.Errorf("GenerateEncParams() ssecurity = %q, want %q", result["ssecurity"], ssecurity)
	}
	if result["_nonce"] != nonce {
		t.Errorf("GenerateEncParams() _nonce = %q, want %q", result["_nonce"], nonce)
	}

	// Original data should be encrypted (not plaintext)
	if result["data"] == `{"did":"12345"}` {
		t.Error("GenerateEncParams() data not encrypted")
	}
}

func TestDecrypt(t *testing.T) {
	ssecurity := "dGVzdHNlY3VyaXR5MTIz"
	nonce := "dGVzdG5vbmNlNDU2"
	plaintext := "hello world"

	// Compute signed nonce
	signedNonce, _ := SignedNonce(ssecurity, nonce)

	// Encrypt
	encrypted, _ := EncryptRC4(signedNonce, plaintext)

	// Decrypt
	decrypted, err := Decrypt(ssecurity, nonce, encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}
