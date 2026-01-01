// +build integration

package client

import (
	"context"
	"fmt"
	"testing"
)

// IntegrationTestConfig holds configuration for integration tests
type IntegrationTestConfig struct {
	Host     string
	Username string
	Password string
}

// defaultIntegrationConfig returns the default configuration for real router testing
func defaultIntegrationConfig() *IntegrationTestConfig {
	return &IntegrationTestConfig{
		Host:     "192.168.1.1",
		Username: "admin",
		Password: "default",
	}
}

// TestIntegration_Login tests authentication against a real router
// Run with: go test -v -tags=integration -run TestIntegration_Login
func TestIntegration_Login(t *testing.T) {
	config := defaultIntegrationConfig()
	
	t.Logf("Testing login to %s with user %s", config.Host, config.Username)

	opts := &Options{
		Auth: fmt.Sprintf("%s:%s", config.Username, config.Password),
		Host: config.Host,
	}

	client, err := NewSMSClient(opts)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.connect(ctx)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if client.sessionID == "" {
		t.Error("Session ID is empty after login")
	}
	if client.tokenID == "" {
		t.Error("Token ID is empty after login")
	}

	t.Logf("✓ Login successful!")
	t.Logf("✓ Session ID: %s", client.sessionID)
	t.Logf("✓ Token ID: %s", client.tokenID)
}

// TestIntegration_ListSMS tests listing SMS messages from the inbox
// Run with: go test -v -tags=integration -run TestIntegration_ListSMS
func TestIntegration_ListSMS(t *testing.T) {
	config := defaultIntegrationConfig()
	
	t.Logf("Testing SMS list on %s", config.Host)

	opts := &Options{
		Auth: fmt.Sprintf("%s:%s", config.Username, config.Password),
		Host: config.Host,
	}

	client, err := NewSMSClient(opts)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.List(ctx, "inbox")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if resp == nil {
		t.Error("Response is nil")
	}

	if resp.Error != 0 {
		t.Logf("⚠ Warning: Router returned error code: %d", resp.Error)
	}

	t.Logf("✓ SMS List successful!")
	t.Logf("✓ Message count: %d", len(resp.Data))
	t.Logf("✓ Error code: %d (0 = no error)", resp.Error)
	
	for i, msg := range resp.Data {
		t.Logf("  [%d] From=%s, Content=%s, Time=%v", i, msg.From, msg.Content, msg.RecvTime)
	}
}

// TestIntegration_ReadSMS tests reading a specific SMS message
// Run with: go test -v -tags=integration -run TestIntegration_ReadSMS
func TestIntegration_ReadSMS(t *testing.T) {
	config := defaultIntegrationConfig()
	
	t.Logf("Testing SMS read on %s", config.Host)

	opts := &Options{
		Auth: fmt.Sprintf("%s:%s", config.Username, config.Password),
		Host: config.Host,
	}

	client, err := NewSMSClient(opts)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	
	// First list to get an existing message
	listResp, err := client.List(ctx, "inbox")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(listResp.Data) == 0 {
		t.Skip("No SMS messages in inbox to test read")
	}

	// Read the first message
	firstIndex := listResp.Data[0].Index
	t.Logf("Reading message at index %d", firstIndex)

	readResp, err := client.Read(ctx, "inbox", firstIndex)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if readResp == nil {
		t.Error("Response is nil")
	}

	if readResp.Error != 0 {
		t.Logf("⚠ Warning: Router returned error code: %d", readResp.Error)
	}

	t.Logf("✓ SMS Read successful!")
	t.Logf("✓ Error code: %d (0 = no error)", readResp.Error)
	
	if len(readResp.Data) > 0 {
		msg := readResp.Data[0]
		t.Logf("  From: %s", msg.From)
		t.Logf("  Content: %s", msg.Content)
		t.Logf("  Received: %v", msg.RecvTime)
	}
}

// TestIntegration_EncryptionParameters tests the encryption parameter retrieval
// Run with: go test -v -tags=integration -run TestIntegration_EncryptionParameters
func TestIntegration_EncryptionParameters(t *testing.T) {
	config := defaultIntegrationConfig()
	
	t.Logf("Testing encryption parameter retrieval from %s", config.Host)

	opts := &Options{
		Auth: fmt.Sprintf("%s:%s", config.Username, config.Password),
		Host: config.Host,
	}

	client, err := NewSMSClient(opts)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	
	// Attempt to fetch encryption parameters via connect
	err = client.connect(ctx)
	if err != nil {
		t.Logf("Connect failed, but may still have params: %v", err)
	}

	// Check if encryption was set up
	if client.enc == nil {
		t.Error("Encryption not initialized")
	} else {
		t.Logf("✓ Encryption initialized successfully")
		t.Logf("  RSA configured: %v", client.enc.rsa != nil)
		t.Logf("  AES configured: %v", client.enc.aes != nil)
		t.Logf("  Sequence: %d", client.enc.seq)
		t.Logf("  Hash: %s", client.enc.hash)
		t.Logf("  AES Key String: %s", client.enc.aesKeyStr)
	}
}

// TestIntegration_FullWorkflow tests the complete workflow
// Run with: go test -v -tags=integration -run TestIntegration_FullWorkflow
func TestIntegration_FullWorkflow(t *testing.T) {
	config := defaultIntegrationConfig()
	
	t.Logf("Testing full SMS workflow on %s", config.Host)

	opts := &Options{
		Auth: fmt.Sprintf("%s:%s", config.Username, config.Password),
		Host: config.Host,
	}

	client, err := NewSMSClient(opts)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Step 1: List messages
	t.Log("\nStep 1: Listing SMS messages...")
	listResp, err := client.List(ctx, "inbox")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if listResp.Error != 0 {
		t.Logf("⚠ List returned error code: %d (expected 0)", listResp.Error)
	} else {
		t.Logf("✓ List successful with error code 0")
	}

	t.Logf("  Found %d messages", len(listResp.Data))

	if len(listResp.Data) > 0 {
		// Step 2: Read first message
		t.Log("\nStep 2: Reading first SMS message...")
		firstIndex := listResp.Data[0].Index
		readResp, err := client.Read(ctx, "inbox", firstIndex)
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}

		if readResp.Error != 0 {
			t.Logf("⚠ Read returned error code: %d (expected 0)", readResp.Error)
		} else {
			t.Logf("✓ Read successful with error code 0")
		}

		if len(readResp.Data) > 0 {
			msg := readResp.Data[0]
			t.Logf("  From: %s", msg.From)
			t.Logf("  Content: %s", msg.Content)
		}
	}

	t.Log("\n✓ Full workflow completed successfully")
}

// TestIntegration_CookieHandling verifies proper cookie storage and use
// Run with: go test -v -tags=integration -run TestIntegration_CookieHandling
func TestIntegration_CookieHandling(t *testing.T) {
	config := defaultIntegrationConfig()
	
	t.Logf("Testing cookie handling on %s", config.Host)

	opts := &Options{
		Auth: fmt.Sprintf("%s:%s", config.Username, config.Password),
		Host: config.Host,
	}

	client, err := NewSMSClient(opts)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Verify session cookie is set
	if client.sessionID == "" {
		t.Error("Session ID not set after connect")
	} else {
		t.Logf("✓ Session ID obtained: %s", client.sessionID)
	}

	// Verify token ID is set
	if client.tokenID == "" {
		t.Error("Token ID not set after connect")
	} else {
		t.Logf("✓ Token ID obtained: %s", client.tokenID)
	}

	// Verify cookie jar has cookies
	if client.httpClient.Jar != nil {
		t.Logf("✓ HTTP client has cookie jar configured")
	}
}
