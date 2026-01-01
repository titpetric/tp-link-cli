package client

import (
	"encoding/base64"
	"strings"
	"testing"
)

// TestAESEncryptionVectors tests AES encryption against known test vectors
// Based on Python implementation test cases
func TestAESEncryptionVectors(t *testing.T) {
	tests := []struct {
		name        string
		keyStr      string
		ivStr       string
		plaintext   string
		expectedB64 string
	}{
		{
			name:        "admin/default login",
			keyStr:      "1767278241989203",
			ivStr:       "1767278241988901",
			plaintext:   "admin\ndefault",
			expectedB64: "u7kfzPnA2T4X4ZJCrUPDbA==",
		},
		{
			name:        "python test vector",
			keyStr:      "1609088003850494",
			ivStr:       "1609088003850873",
			plaintext:   "This is a text message encrypted by my own AES encoding algorithm. Enjoy!",
			expectedB64: "u1LrU4tN+9fr8kZgrIvGJKd76vhHjFpT29hHxcS0W7U9Rxe+YHmi28CPFiz+fkGYsyuj3X9XKYxj0zZS82hhCNjWlD7BUeWLbLcLVCvUMsE=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aes := NewAES()
			aes.SetKeyFromNumeric(tt.keyStr, tt.ivStr)

			ciphertext := aes.Encrypt(tt.plaintext)

			if ciphertext != tt.expectedB64 {
				t.Errorf("Encryption mismatch\nGot:      %s\nExpected: %s", ciphertext, tt.expectedB64)
			}

			// Test decryption
			decrypted, err := aes.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Decryption mismatch\nGot:      %q\nExpected: %q", decrypted, tt.plaintext)
			}

			t.Logf("✓ Encryption/decryption successful")
			t.Logf("  Plaintext length: %d", len(tt.plaintext))
			t.Logf("  Ciphertext length: %d", len(ciphertext))
		})
	}
}

// TestProtocolFrameEncryption tests encryption of protocol frames
func TestProtocolFrameEncryption(t *testing.T) {
	// Create encryption instance with test keys
	enc := NewEncryption()
	enc.SetRSAKey(
		"E66FDAC84695316901FD021515E50289660E7EAD252CAAC5B56FFC1332B4BEF6FAB44C01A2510C3053C1CC259D9983FB1719F9F9FA7B96AE65860BDBA97AC4C3",
		"010001",
	)
	enc.SetHash("admin", "default")
	enc.SetSeq(585322885)
	enc.SetAESKey("1767278241989203", "1767278241988901")

	// Test a simple protocol frame
	dataFrame := "1\r\n[LTE_SMS_RECVMSGBOX#0,0,0,0,0,0#0,0,0,0,0,0]0,1\r\nPageNumber=1\r\n"

	result := enc.AESEncrypt(dataFrame, false)

	if result.Data == "" {
		t.Error("Encrypted data is empty")
	}

	if result.Sign == "" {
		t.Error("RSA signature is empty")
	}

	// Verify it's valid base64
	_, err := base64.StdEncoding.DecodeString(result.Data)
	if err != nil {
		t.Errorf("Encrypted data is not valid base64: %v", err)
	}

	// Verify it's valid hex for RSA signature
	if len(result.Sign)%2 != 0 {
		t.Error("RSA signature length is odd (should be even hex)")
	}

	t.Logf("✓ Protocol frame encryption successful")
	t.Logf("  Frame length: %d", len(dataFrame))
	t.Logf("  Encrypted data length: %d", len(result.Data))
	t.Logf("  Signature length: %d hex chars", len(result.Sign))
}

// TestSignatureFormat tests that signatures have correct format
func TestSignatureFormat(t *testing.T) {
	enc := NewEncryption()
	enc.SetRSAKey(
		"E66FDAC84695316901FD021515E50289660E7EAD252CAAC5B56FFC1332B4BEF6FAB44C01A2510C3053C1CC259D9983FB1719F9F9FA7B96AE65860BDBA97AC4C3",
		"010001",
	)
	enc.SetHash("admin", "default")
	enc.SetSeq(100)
	enc.SetAESKey("1767278241989203", "1767278241988901")

	// Test login signature format
	result := enc.AESEncrypt("test", true)

	// For login, signature should include: key=...&iv=...&h=...&s=...
	signature := result.Sign

	t.Logf("Login signature: %s", signature)

	// Verify signature is hex
	for _, c := range signature {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Errorf("Invalid character in signature: %c", c)
		}
	}

	// Test non-login signature format
	result2 := enc.AESEncrypt("test", false)
	signature2 := result2.Sign

	t.Logf("Non-login signature: %s", signature2)

	// Both should be hex strings
	for _, c := range signature2 {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Errorf("Invalid character in non-login signature: %c", c)
		}
	}

	t.Log("✓ Signature format validation passed")
}

// TestDataLengthConsistency verifies encrypted data length is consistent
func TestDataLengthConsistency(t *testing.T) {
	aes := NewAES()
	aes.SetKeyFromNumeric("1767278241989203", "1767278241988901")

	testCases := []string{
		"a",
		"admin",
		"admin\ndefault",
		"This is a longer test string that should be padded",
		"a" + strings.Repeat("b", 15), // Exactly 16 bytes
		"a" + strings.Repeat("b", 16), // 17 bytes - should pad to 32
	}

	for i, plaintext := range testCases {
		ciphertext := aes.Encrypt(plaintext)

		// Verify base64 encoding
		decoded, err := base64.StdEncoding.DecodeString(ciphertext)
		if err != nil {
			t.Errorf("Case %d: Invalid base64: %v", i, err)
		}

		// All blocks should be multiple of 16 (AES block size)
		if len(decoded)%16 != 0 {
			t.Errorf("Case %d: Decoded length %d is not multiple of 16", i, len(decoded))
		}

		t.Logf("Case %d: plaintext_len=%d, decoded_len=%d, base64_len=%d",
			i, len(plaintext), len(decoded), len(ciphertext))
	}

	t.Log("✓ Data length consistency verified")
}
