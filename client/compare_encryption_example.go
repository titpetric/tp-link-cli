package client

import (
	"fmt"
)

// ExampleCompareEncryption demonstrates encryption comparison with Python.
func ExampleCompareEncryption() {
	// Use fixed key/IV like Python would generate from a known timestamp
	// From Python output: key=1767278241989203&iv=1767278241988901

	enc := NewEncryption()

	// Set RSA key
	nn := "E66FDAC84695316901FD021515E50289660E7EAD252CAAC5B56FFC1332B4BEF6FAB44C01A2510C3053C1CC259D9983FB1719F9F9FA7B96AE65860BDBA97AC4C3"
	ee := "010001"
	enc.SetRSAKey(nn, ee)

	// Set hash
	enc.SetHash("admin", "default")

	// Set sequence
	enc.SetSeq(585322885)

	// Manually set AES key to match Python's (numerically derived)
	// Key string from Python: key=1767278241989203&iv=1767278241988901
	enc.SetAESKey("1767278241989203", "1767278241988901")

	fmt.Printf("AES Key String: %s\n", enc.GetAESKeyString())

	// Encrypt auth data
	authData := "admin\ndefault"
	result := enc.AESEncrypt(authData, true)

	fmt.Printf("\nEncrypted Data: %s\n", result.Data)
	fmt.Printf("Expected Data: u7kfzPnA2T4X4ZJCrUPDbA==\n")
	fmt.Printf("Match: %v\n", result.Data == "u7kfzPnA2T4X4ZJCrUPDbA==")

	fmt.Printf("\nRSA Sign: %s\n", result.Sign)
}
