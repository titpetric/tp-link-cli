package client

import (
	"fmt"
	"testing"
)

// TestEncryptionMatch tests that encryption matches Python implementation
func TestEncryptionMatch(t *testing.T) {
	enc := NewEncryption()

	// Set RSA key - this is fixed for test
	nn := "E66FDAC84695316901FD021515E50289660E7EAD252CAAC5B56FFC1332B4BEF6FAB44C01A2510C3053C1CC259D9983FB1719F9F9FA7B96AE65860BDBA97AC4C3"
	ee := "010001"
	enc.SetRSAKey(nn, ee)

	// Set hash
	enc.SetHash("admin", "default")

	// Set sequence
	enc.SetSeq(585322885)

	// Manually set AES key to match Python's (numerically derived)
	enc.SetAESKey("1767278241989203", "1767278241988901")

	aesKeyStr := enc.GetAESKeyString()
	t.Logf("AES Key String: %s", aesKeyStr)

	expectedKeyStr := "key=1767278241989203&iv=1767278241988901"
	if aesKeyStr != expectedKeyStr {
		t.Errorf("AES Key String mismatch. Got: %s, Expected: %s", aesKeyStr, expectedKeyStr)
	}

	// Encrypt auth data
	authData := "admin\ndefault"
	result := enc.AESEncrypt(authData, true)

	t.Logf("Encrypted Data: %s", result.Data)
	t.Logf("Expected Data: u7kfzPnA2T4X4ZJCrUPDbA==")

	if result.Data != "u7kfzPnA2T4X4ZJCrUPDbA==" {
		t.Errorf("Encryption mismatch. Got: %s", result.Data)
	}

	t.Logf("RSA Sign: %s", result.Sign)
}

// TestAESKeyStringParsing tests key derivation from numeric strings
func TestAESKeyStringParsing(t *testing.T) {
	aes := NewAES()

	// Test with known values
	keyStr := "1767278241989203"
	ivStr := "1767278241988901"
	aes.SetKeyFromNumeric(keyStr, ivStr)

	result := aes.GetKeyString()
	expected := "key=1767278241989203&iv=1767278241988901"

	if result != expected {
		t.Errorf("Key string format mismatch. Got: %s, Expected: %s", result, expected)
	}

	// Test encryption with known plaintext
	plaintext := "admin\ndefault"
	ciphertext := aes.Encrypt(plaintext)

	t.Logf("Plaintext: %q", plaintext)
	t.Logf("Ciphertext: %s", ciphertext)

	// The ciphertext length should be consistent
	t.Logf("Ciphertext length: %d", len(ciphertext))

	// Decryption test
	decrypted, err := aes.Decrypt(ciphertext)
	if err != nil {
		t.Errorf("Decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decryption mismatch. Got: %q, Expected: %q", decrypted, plaintext)
	}

	t.Log("AES round-trip encryption successful")
}

// TestUtf8ParseToBytes validates the key/IV byte conversion
func TestUtf8ParseToBytes(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		minLen int
		maxLen int
	}{
		{
			name:   "16-digit key",
			input:  "1767278241989203",
			minLen: 16,
			maxLen: 16,
		},
		{
			name:   "16-digit iv",
			input:  "1767278241988901",
			minLen: 16,
			maxLen: 16,
		},
		{
			name:   "short string",
			input:  "short",
			minLen: 16, // Should be padded to 16
			maxLen: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := utf8ParseToBytes(tt.input)
			if len(bytes) < tt.minLen || len(bytes) > tt.maxLen {
				t.Errorf("Length out of range. Got: %d, Expected: %d-%d", len(bytes), tt.minLen, tt.maxLen)
			}

			// Display byte values for debugging
			fmt.Printf("Input: %q -> Bytes: %v\n", tt.input, bytes)
		})
	}
}
