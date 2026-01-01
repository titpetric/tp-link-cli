//go:build integration

package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/titpetric/tp-link-cli/client"
)

func NewClientOptions() *client.Options {
	auth := os.Getenv("TP_LINK_CLI_AUTH")
	if auth == "" {
		auth = "admin:admin"
	}

	host := os.Getenv("TP_LINK_CLI_HOST")
	if host == "" {
		host = "192.168.1.1"
	}
	return &client.Options{
		Auth: auth,
		Host: host,
	}
}

// TestLoginIntegration tests login against real router at 192.168.1.1
// This is an integration test that requires access to the router
func TestLoginIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := client.NewSMSClient(NewClientOptions())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if client.SessionID == "" {
		t.Error("Session ID is empty after login")
	}
	if client.TokenID == "" {
		t.Error("Token ID is empty after login")
	}

	t.Logf("Login successful!")
	t.Logf("Session ID: %s", client.SessionID)
	t.Logf("Token ID: %s", client.TokenID)
}

// TestSMSListIntegration tests listing SMS messages from the router
func TestSMSListIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := client.NewSMSClient(NewClientOptions())
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

	t.Logf("SMS List successful!")
	t.Logf("Message count: %d", len(resp.Data))
	for i, msg := range resp.Data {
		t.Logf("Message %d: From=%s, Content=%s", i, msg.From, msg.Content)
	}
}
