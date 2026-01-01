package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

// AES handles AES-128-CBC encryption/decryption
type AES struct {
	key       []byte
	iv        []byte
	keyNum    int64  // numeric key for signature (like Python implementation)
	ivNum     int64  // numeric iv for signature
	keyNumStr string // original key string (preserves leading zeros)
	ivNumStr  string // original iv string (preserves leading zeros)
}

// NewAES creates a new AES cipher with random key and IV
func NewAES() *AES {
	a := &AES{}
	a.genKey()
	return a
}

func (a *AES) genKey() {
	// Generate numeric key and IV based on timestamp (like Python)
	now := time.Now()
	micros := now.UnixMicro()

	buf := make([]byte, 4)
	rand.Read(buf)
	randValKey := int64(int(buf[0]) % 1000)

	rand.Read(buf)
	randValIV := int64(int(buf[0]) % 1000)

	keyNum := micros + randValKey
	ivNum := micros + randValIV

	// Convert to 16-character numeric strings
	keyStr := fmt.Sprintf("%d", keyNum)[:16]
	ivStr := fmt.Sprintf("%d", ivNum)[:16]

	// But for the actual AES key, convert via utf8 encoding like Python does
	key := utf8ParseToBytes(keyStr)
	iv := utf8ParseToBytes(ivStr)

	a.key = key
	a.iv = iv
	a.keyNum = keyNum
	a.ivNum = ivNum
	a.keyNumStr = keyStr
	a.ivNumStr = ivStr
}

// utf8ParseToBytes converts a string to bytes for AES, matching Python's utf8Parse
func utf8ParseToBytes(s string) []byte {
	var result []byte
	for i := 0; i < len(s) && len(result) < 16; i++ {
		result = append(result, byte(s[i]))
	}
	// Pad to 16 bytes
	for len(result) < 16 {
		result = append(result, 0)
	}
	return result
}

// SetKey sets the AES key and IV from bytes
func (a *AES) SetKey(key, iv []byte) {
	a.key = key
	a.iv = iv
}

// SetKeyFromHex sets the AES key and IV from hex strings
func (a *AES) SetKeyFromHex(keyHex, ivHex string) error {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return err
	}
	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		return err
	}
	a.key = key
	a.iv = iv
	return nil
}

// SetKeyFromNumeric sets the AES key and IV from numeric strings (like Python uses)
func (a *AES) SetKeyFromNumeric(keyNum, ivNum string) {
	a.key = utf8ParseToBytes(keyNum)
	a.iv = utf8ParseToBytes(ivNum)
	a.keyNumStr = keyNum
	a.ivNumStr = ivNum
	// Parse the numeric values for comparison/validation
	if v, err := strconv.ParseInt(keyNum, 10, 64); err == nil {
		a.keyNum = v
	}
	if v, err := strconv.ParseInt(ivNum, 10, 64); err == nil {
		a.ivNum = v
	}
}

// GetKeyString returns the key in the format used by the router (numeric format like Python)
func (a *AES) GetKeyString() string {
	// Use original string values if available to preserve leading zeros
	if a.keyNumStr != "" && a.ivNumStr != "" {
		return fmt.Sprintf("key=%s&iv=%s", a.keyNumStr, a.ivNumStr)
	}
	return fmt.Sprintf("key=%d&iv=%d", a.keyNum, a.ivNum)
}

// Encrypt encrypts plaintext using AES-128-CBC
func (a *AES) Encrypt(plaintext string) string {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return ""
	}

	plainBytes := []byte(plaintext)
	// PKCS7 padding
	padLen := aes.BlockSize - (len(plainBytes) % aes.BlockSize)
	for i := 0; i < padLen; i++ {
		plainBytes = append(plainBytes, byte(padLen))
	}

	mode := cipher.NewCBCEncrypter(block, a.iv)
	ciphertext := make([]byte, len(plainBytes))
	mode.CryptBlocks(ciphertext, plainBytes)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

// Decrypt decrypts base64-encoded ciphertext using AES-128-CBC
func (a *AES) Decrypt(ciphertext string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, a.iv)
	plaintext := make([]byte, len(cipherBytes))
	mode.CryptBlocks(plaintext, cipherBytes)

	// Remove PKCS7 padding
	padLen := int(plaintext[len(plaintext)-1])
	plaintext = plaintext[:len(plaintext)-padLen]

	return string(plaintext), nil
}

// RSAKey handles RSA 512-bit operations
type RSAKey struct {
	n *big.Int // modulus
	e int64    // exponent
}

// NewRSAKey creates a new RSA key with given public key parameters
func NewRSAKey(nHex string, eHex string) (*RSAKey, error) {
	n := new(big.Int)
	n.SetString(nHex, 16)
	e, err := strconv.ParseInt(eHex, 16, 64)
	if err != nil {
		return nil, err
	}
	return &RSAKey{n: n, e: e}, nil
}

// Encrypt performs RSA encryption on plaintext
func (r *RSAKey) Encrypt(plaintext string) string {
	m := r.noPadding(plaintext)
	if m == nil {
		return ""
	}

	c := new(big.Int)
	c.Exp(m, big.NewInt(r.e), r.n)

	hexStr := fmt.Sprintf("%x", c)
	// Pad to the full RSA modulus size in hex
	// RSA-512 = 64 bytes = 128 hex chars
	expectedLen := (r.n.BitLen() + 3) / 4 // Convert bits to hex chars
	for len(hexStr) < expectedLen {
		hexStr = "0" + hexStr
	}
	return hexStr
}

// noPadding converts string to big.Int with UTF-8 encoding
func (r *RSAKey) noPadding(s string) *big.Int {
	blockSize := (r.n.BitLen() + 7) >> 3
	byteArray := make([]byte, blockSize)

	i := 0
	j := 0
	for i < len(s) && j < blockSize {
		ch := rune(s[i])
		if ch < 128 {
			byteArray[j] = byte(ch)
			j++
		} else if ch < 2048 {
			byteArray[j] = byte((ch & 63) | 128)
			j++
			byteArray[j] = byte((ch >> 6) | 192)
			j++
		} else {
			byteArray[j] = byte((ch & 63) | 128)
			j++
			byteArray[j] = byte(((ch >> 6) & 63) | 128)
			j++
			byteArray[j] = byte((ch >> 12) | 224)
			j++
		}
		i++
	}

	return new(big.Int).SetBytes(byteArray)
}

// Encryption manages both AES and RSA encryption
type Encryption struct {
	aes       *AES
	rsa       *RSAKey
	seq       int
	aesKeyStr string
	hash      string
}

// NewEncryption creates a new Encryption manager
func NewEncryption() *Encryption {
	return &Encryption{
		aes: NewAES(),
	}
}

// SetHash computes hash from username and password
func (e *Encryption) SetHash(username, password string) {
	h := md5.Sum([]byte(username + password))
	e.hash = fmt.Sprintf("%x", h)
}

// SetSeq sets the sequence number
func (e *Encryption) SetSeq(seq int) {
	e.seq = seq
}

// SetRSAKey configures the RSA public key
func (e *Encryption) SetRSAKey(nHex, eHex string) error {
	rsa, err := NewRSAKey(nHex, eHex)
	if err != nil {
		return err
	}
	e.rsa = rsa
	return nil
}

// GenAESKey generates a new AES key and IV
func (e *Encryption) GenAESKey() {
	e.aes.genKey()
	e.aesKeyStr = e.aes.GetKeyString()
}

// GetAESKeyString returns the AES key string for authentication
func (e *Encryption) GetAESKeyString() string {
	return e.aesKeyStr
}

// SetAESKey sets the AES key and IV from numeric strings
func (e *Encryption) SetAESKey(key, iv string) error {
	e.aes.SetKeyFromNumeric(key, iv)
	e.aesKeyStr = e.aes.GetKeyString()
	return nil
}

// AESEncryptResult holds encrypted data and signature
type AESEncryptResult struct {
	Data string
	Sign string
}

// AESEncrypt encrypts data with AES and signs with RSA
func (e *Encryption) AESEncrypt(data string, isLogin bool) AESEncryptResult {
	encrypted := e.aes.Encrypt(data)

	signature := ""
	if isLogin {
		signature = e.aesKeyStr + "&"
	}

	dataLen := len(encrypted)
	signature += fmt.Sprintf("h=%s&s=%d", e.hash, e.seq+dataLen)

	fmt.Fprintf(os.Stderr, "DEBUG AESEncrypt: plaintext: %q\n", data)
	fmt.Fprintf(os.Stderr, "DEBUG AESEncrypt: encrypted (base64): %s\n", encrypted)
	fmt.Fprintf(os.Stderr, "DEBUG AESEncrypt: encrypted length: %d\n", dataLen)
	fmt.Fprintf(os.Stderr, "DEBUG AESEncrypt: aes_key_str: %s\n", e.aesKeyStr)
	fmt.Fprintf(os.Stderr, "DEBUG AESEncrypt: hash: %s\n", e.hash)
	fmt.Fprintf(os.Stderr, "DEBUG AESEncrypt: seq: %d\n", e.seq)
	fmt.Fprintf(os.Stderr, "DEBUG AESEncrypt: data to sign: %s\n", signature)

	signed := e.rsa.Encrypt(signature)
	fmt.Fprintf(os.Stderr, "DEBUG AESEncrypt: rsa signature: %s\n", signed)

	return AESEncryptResult{
		Data: encrypted,
		Sign: signed,
	}
}

// AESDecrypt decrypts AES-encrypted data
func (e *Encryption) AESDecrypt(data string) (string, error) {
	return e.aes.Decrypt(data)
}

// ParseEncryptionParams extracts encryption parameters from getParm response
func ParseEncryptionParams(response string) (ee, nn, seq string, err error) {
	// Extract ee (exponent)
	eeStart := strings.Index(response, `ee="`)
	if eeStart == -1 {
		return "", "", "", fmt.Errorf("ee parameter not found")
	}
	eeStart += 4
	eeEnd := strings.Index(response[eeStart:], `"`)
	if eeEnd == -1 {
		return "", "", "", fmt.Errorf("ee parameter malformed")
	}
	ee = response[eeStart : eeStart+eeEnd]

	// Extract nn (modulus)
	nnStart := strings.Index(response, `nn="`)
	if nnStart == -1 {
		return "", "", "", fmt.Errorf("nn parameter not found")
	}
	nnStart += 4
	nnEnd := strings.Index(response[nnStart:], `"`)
	if nnEnd == -1 {
		return "", "", "", fmt.Errorf("nn parameter malformed")
	}
	nn = response[nnStart : nnStart+nnEnd]

	// Extract seq (sequence)
	seqStart := strings.Index(response, `seq="`)
	if seqStart == -1 {
		return "", "", "", fmt.Errorf("seq parameter not found")
	}
	seqStart += 5
	seqEnd := strings.Index(response[seqStart:], `"`)
	if seqEnd == -1 {
		return "", "", "", fmt.Errorf("seq parameter malformed")
	}
	seq = response[seqStart : seqStart+seqEnd]

	return ee, nn, seq, nil
}
