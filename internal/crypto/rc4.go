package crypto

import (
	"bytes"
	"compress/gzip"
	"crypto/rc4"
	"encoding/base64"
	"io"
)

// EncryptRC4 encrypts a plaintext string using RC4 with drop-1024.
// Password is base64-encoded, payload is plaintext string.
// Returns base64-encoded ciphertext.
func EncryptRC4(password, payload string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return "", err
	}

	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Drop first 1024 bytes (RC4 security mitigation)
	drop := make([]byte, 1024)
	cipher.XORKeyStream(drop, drop)

	// Encrypt payload
	dst := make([]byte, len(payload))
	cipher.XORKeyStream(dst, []byte(payload))

	return base64.StdEncoding.EncodeToString(dst), nil
}

// DecryptRC4 decrypts a base64-encoded ciphertext using RC4 with drop-1024.
// Password is base64-encoded, payload is base64-encoded ciphertext.
// Returns decrypted bytes.
func DecryptRC4(password, payload string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return nil, err
	}

	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Drop first 1024 bytes (RC4 security mitigation)
	drop := make([]byte, 1024)
	cipher.XORKeyStream(drop, drop)

	// Decode and decrypt
	encrypted, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}

	dst := make([]byte, len(encrypted))
	cipher.XORKeyStream(dst, encrypted)

	return dst, nil
}

// Decrypt decrypts a response payload using RC4 with gzip fallback.
func Decrypt(ssecurity, nonce, payload string) (string, error) {
	signedNonce, err := SignedNonce(ssecurity, nonce)
	if err != nil {
		return "", err
	}

	decrypted, err := DecryptRC4(signedNonce, payload)
	if err != nil {
		return "", err
	}

	// Try UTF-8 decode first
	if result, err := utf8Decode(decrypted); err == nil {
		return result, nil
	}

	// Fall back to gzip decompression
	return gzipDecode(decrypted)
}

func utf8Decode(data []byte) (string, error) {
	// Simple UTF-8 validation: check if all bytes are valid UTF-8
	for i := 0; i < len(data); {
		if data[i] <= 0x7F {
			i++
		} else if data[i] >= 0xC0 && data[i] <= 0xDF {
			if i+1 >= len(data) || data[i+1] < 0x80 || data[i+1] > 0xBF {
				return "", io.ErrUnexpectedEOF
			}
			i += 2
		} else if data[i] >= 0xE0 && data[i] <= 0xEF {
			if i+2 >= len(data) || data[i+1] < 0x80 || data[i+1] > 0xBF || data[i+2] < 0x80 || data[i+2] > 0xBF {
				return "", io.ErrUnexpectedEOF
			}
			i += 3
		} else if data[i] >= 0xF0 && data[i] <= 0xF7 {
			if i+3 >= len(data) || data[i+1] < 0x80 || data[i+1] > 0xBF || data[i+2] < 0x80 || data[i+2] > 0xBF || data[i+3] < 0x80 || data[i+3] > 0xBF {
				return "", io.ErrUnexpectedEOF
			}
			i += 4
		} else {
			return "", io.ErrUnexpectedEOF
		}
	}
	return string(data), nil
}

func gzipDecode(data []byte) (string, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	result, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
