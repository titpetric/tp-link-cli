package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAESEncryptDecrypt(t *testing.T) {
	aes := NewAES()

	plaintext := "Hello, World!"
	encrypted := aes.Encrypt(plaintext)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := aes.Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAESKeyString(t *testing.T) {
	aes := NewAES()
	keyStr := aes.GetKeyString()

	assert.Contains(t, keyStr, "key=")
	assert.Contains(t, keyStr, "iv=")
	assert.Contains(t, keyStr, "&")
}

func TestParseEncryptionParams(t *testing.T) {
	response := `ee="10001" nn="ABCD1234" seq="123" other="value"`

	ee, nn, seq, err := ParseEncryptionParams(response)
	assert.NoError(t, err)
	assert.Equal(t, "10001", ee)
	assert.Equal(t, "ABCD1234", nn)
	assert.Equal(t, "123", seq)
}

func TestParseEncryptionParamsError(t *testing.T) {
	response := `invalid response`

	_, _, _, err := ParseEncryptionParams(response)
	assert.Error(t, err)
}

func TestRSAKeyEncrypt(t *testing.T) {
	// Using test values - note: real RSA needs proper key setup
	rsa, err := NewRSAKey("C77FFBF20F381C2B8050FD9BAA3E25D4", "10001")
	assert.NoError(t, err)

	plaintext := "test"
	encrypted := rsa.Encrypt(plaintext)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)
}

func TestEncryptionAESEncryptDecrypt(t *testing.T) {
	enc := NewEncryption()

	plaintext := "test data"
	encrypted := enc.aes.Encrypt(plaintext)
	assert.NotEmpty(t, encrypted)

	decrypted, err := enc.AESDecrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptionSetAESKey(t *testing.T) {
	enc := NewEncryption()

	// Test with numeric strings (16 digits each)
	err := enc.SetAESKey("0123456789012345", "5432109876543210")
	assert.NoError(t, err)

	keyStr := enc.GetAESKeyString()
	assert.Contains(t, keyStr, "0123456789012345")
	assert.Contains(t, keyStr, "5432109876543210")
}
