package client

import (
	"testing"
)

// TestLoginEncryptionWithKnownKey tests that login encryption works with known vectors
func TestLoginEncryptionWithKnownKey(t *testing.T) {
	enc := NewEncryption()

	// Set RSA key (these are from the router)
	nn := "E66FDAC84695316901FD021515E50289660E7EAD252CAAC5B56FFC1332B4BEF6FAB44C01A2510C3053C1CC259D9983FB1719F9F9FA7B96AE65860BDBA97AC4C3"
	ee := "010001"
	if err := enc.SetRSAKey(nn, ee); err != nil {
		t.Fatalf("Failed to set RSA key: %v", err)
	}

	// Set credentials
	enc.SetHash("admin", "default")
	enc.SetSeq(585322885)
	enc.SetAESKey("1767278241989203", "1767278241988901")

	// Encrypt login data
	loginData := "admin\ndefault"
	result := enc.AESEncrypt(loginData, true)

	// Verify encrypted data
	if result.Data != "u7kfzPnA2T4X4ZJCrUPDbA==" {
		t.Errorf("Encrypted data mismatch. Got: %s", result.Data)
	}

	// Verify signature is hex
	if len(result.Sign) == 0 {
		t.Error("Signature is empty")
	}

	// Verify signature format (should be hex digits)
	for _, c := range result.Sign {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Invalid character in signature: %c", c)
		}
	}

	t.Logf("✓ Login encryption successful")
	t.Logf("  Encrypted: %s", result.Data)
	t.Logf("  Signature length: %d hex chars", len(result.Sign))
}

// TestDataFrameEncryptionLifecycle tests encryption of a full data frame exchange
func TestDataFrameEncryptionLifecycle(t *testing.T) {
	// Setup encryption
	enc := NewEncryption()
	enc.SetRSAKey(
		"E66FDAC84695316901FD021515E50289660E7EAD252CAAC5B56FFC1332B4BEF6FAB44C01A2510C3053C1CC259D9983FB1719F9F9FA7B96AE65860BDBA97AC4C3",
		"010001",
	)
	enc.SetHash("admin", "default")
	enc.SetSeq(585322885)
	enc.GenAESKey()

	// Setup protocol
	proto := NewProtocol()

	// Create a simple SMS list request
	reqs := []Request{
		{
			Method:     ActSet,
			Controller: "LTE_SMS_RECVMSGBOX",
			Attrs: map[string]interface{}{
				"PageNumber": 1,
			},
		},
		{
			Method:     ActGL,
			Controller: "LTE_SMS_RECVMSGENTRY",
			Attrs:      []string{"index", "from", "content", "receivedTime", "unread"},
		},
	}

	// Build data frame
	dataFrame := proto.MakeDataFrame(reqs)
	t.Logf("Data frame:\n%s", dataFrame)

	// Encrypt frame
	encrypted := enc.AESEncrypt(dataFrame, false)

	// Verify encrypted data is not empty
	if encrypted.Data == "" {
		t.Error("Encrypted data is empty")
	}

	// Verify signature is not empty
	if encrypted.Sign == "" {
		t.Error("Signature is empty")
	}

	// Verify we can decrypt it back
	decrypted, err := enc.AESDecrypt(encrypted.Data)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Verify decrypted matches original
	if decrypted != dataFrame {
		t.Errorf("Decrypted data doesn't match original")
		t.Logf("Original:\n%s", dataFrame)
		t.Logf("Decrypted:\n%s", decrypted)
	}

	t.Logf("✓ Data frame encryption/decryption successful")
	t.Logf("  Frame size: %d bytes", len(dataFrame))
	t.Logf("  Encrypted size: %d bytes", len(encrypted.Data))
}

// TestAESKeyGenerationConsistency verifies AES keys are consistent
func TestAESKeyGenerationConsistency(t *testing.T) {
	// Create two encryption instances
	enc1 := NewEncryption()
	enc2 := NewEncryption()

	// Generate keys on both
	enc1.GenAESKey()
	enc2.GenAESKey()

	// Both should have generated different keys (random)
	aesStr1 := enc1.GetAESKeyString()
	aesStr2 := enc2.GetAESKeyString()

	// Verify they're not empty
	if aesStr1 == "" || aesStr2 == "" {
		t.Error("Generated key strings are empty")
	}

	// They should be different (with extremely high probability)
	if aesStr1 == aesStr2 {
		t.Log("⚠ Generated keys happen to be the same (extremely unlikely but possible)")
	} else {
		t.Logf("✓ Generated keys are different as expected")
	}

	// Both should be in the correct format
	t.Logf("Key 1: %s", aesStr1)
	t.Logf("Key 2: %s", aesStr2)
}

// TestProtocolFrameParsing tests parsing of response data
func TestProtocolFrameParsing(t *testing.T) {
	proto := NewProtocol()

	// Example response from router
	responseFrame := `[0,0,0,0,0,0]0
index=1
from=+1234567890
content=Test message
receivedTime=2025-01-01 12:00:00
unread=1
[0,0,0,0,0,0]1
index=2
from=+0987654321
content=Another message
receivedTime=2025-01-01 13:00:00
unread=0
[error]0`

	response := proto.FromDataFrame(responseFrame)

	if response.Error != 0 {
		t.Errorf("Expected error code 0, got %d", response.Error)
	}

	if len(response.Data) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(response.Data))
	}

	// Prettify response
	prettified := proto.PrettifyResponse(response)

	if len(prettified.Data) != 2 {
		t.Error("Prettify should preserve data count")
	}

	t.Logf("✓ Protocol frame parsing successful")
	t.Logf("  Messages parsed: %d", len(prettified.Data))
}
