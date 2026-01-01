package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/titpetric/tp-link-cli/client"
	"github.com/titpetric/tp-link-cli/model"

	"github.com/olekukonko/tablewriter"
)

// SMSCommand handles SMS operations.
type SMSCommand struct {
	Auth   string
	Host   string
	JSON   bool
	Folder string
}

// NewSMSCommand will return the environment-filled *SMSCommand.
// The defaults are applied if no environment is provided.
func NewSMSCommand() *SMSCommand {
	auth := os.Getenv("TP_LINK_CLI_AUTH")
	if auth == "" {
		auth = "admin:admin"
	}

	host := os.Getenv("TP_LINK_CLI_HOST")
	if host == "" {
		host = "192.168.1.1"
	}

	return &SMSCommand{
		Auth: auth,
		Host: host,
	}
}

func (c *SMSCommand) ClientOptions() *client.Options {
	return &client.Options{
		Auth: c.Auth,
		Host: c.Host,
	}
}

// ParseArgs parses command-line arguments
func ParseArgs(args []string) (*SMSCommand, string, error) {
	if len(args) == 0 {
		return nil, "", fmt.Errorf("no subcommand provided")
	}

	cmd := NewSMSCommand()

	subcommand := args[0]
	args = args[1:]

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--json" {
			cmd.JSON = true
		} else if len(arg) > 7 && arg[:7] == "--auth=" {
			cmd.Auth = arg[7:]
		} else if len(arg) > 7 && arg[:7] == "--host=" {
			cmd.Host = arg[7:]
		} else if len(arg) > 9 && arg[:9] == "--folder=" {
			cmd.Folder = arg[9:]
		}
	}

	if cmd.Folder == "" {
		cmd.Folder = "inbox"
	}

	return cmd, subcommand, nil
}

// ListSMS lists SMS messages
func (c *SMSCommand) ListSMS(ctx context.Context) error {
	smsClient, err := client.NewSMSClient(c.ClientOptions())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	resp, err := smsClient.List(ctx, c.Folder)
	if err != nil {
		return fmt.Errorf("failed to list SMS: %w", err)
	}

	if resp.Error != 0 {
		return fmt.Errorf("router returned error code: %d", resp.Error)
	}

	if c.JSON {
		return c.outputJSON(resp.Data)
	}
	return c.outputTable(resp.Data)
}

// ReadSMS reads a specific SMS message
func (c *SMSCommand) ReadSMS(ctx context.Context, index int) error {
	smsClient, err := client.NewSMSClient(c.ClientOptions())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	resp, err := smsClient.Read(ctx, c.Folder, index)
	if err != nil {
		return fmt.Errorf("failed to read SMS: %w", err)
	}

	if resp.Error != 0 {
		return fmt.Errorf("router returned error code: %d", resp.Error)
	}

	if c.JSON {
		return c.outputJSON(resp.Data)
	}

	// Output a single message in readable format
	if len(resp.Data) == 0 {
		fmt.Println("No message found")
		return nil
	}

	msg := resp.Data[0]
	fmt.Printf("Index: %d\n", msg.Index)

	if c.Folder == "inbox" {
		fmt.Printf("From: %s\n", msg.From)
		fmt.Printf("Received: %s\n", formatTime(msg.RecvTime))
		fmt.Printf("Unread: %v\n", msg.Unread)
	} else {
		fmt.Printf("To: %s\n", msg.To)
		fmt.Printf("Sent: %s\n", formatTime(msg.SentTime))
	}

	fmt.Printf("\nMessage:\n%s\n", msg.Content)
	return nil
}

// DeleteSMS deletes a specific SMS message by position
func (c *SMSCommand) DeleteSMS(ctx context.Context, index int) error {
	smsClient, err := client.NewSMSClient(c.ClientOptions())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	resp, err := smsClient.Delete(ctx, c.Folder, index)
	if err != nil {
		return fmt.Errorf("failed to delete SMS: %w", err)
	}

	if resp.Error != 0 {
		return fmt.Errorf("router returned error code: %d", resp.Error)
	}

	fmt.Printf("Message at position %d deleted\n", index)
	return nil
}

// SendSMS sends an SMS message to a phone number
func (c *SMSCommand) SendSMS(ctx context.Context, number, message string) error {
	smsClient, err := client.NewSMSClient(c.ClientOptions())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	resp, err := smsClient.Send(ctx, number, message)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	if resp.Error != 0 {
		return fmt.Errorf("router returned error code: %d", resp.Error)
	}

	fmt.Printf("Message sent to %s\n", number)
	return nil
}

// DeleteSMSByID deletes a specific SMS message by its ID (index)
func (c *SMSCommand) DeleteSMSByID(ctx context.Context, msgID int) error {
	smsClient, err := client.NewSMSClient(c.ClientOptions())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// First, get the list to find the position of the message with this ID
	resp, err := smsClient.List(ctx, c.Folder)
	if err != nil {
		return fmt.Errorf("failed to list SMS: %w", err)
	}

	if resp.Error != 0 {
		return fmt.Errorf("router returned error code: %d", resp.Error)
	}

	// Find the message with matching ID
	var position int
	found := false
	for i, msg := range resp.Data {
		if msg.Index == msgID {
			position = i + 1 // Convert to 1-based position
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("message with ID %d not found in %s folder", msgID, c.Folder)
	}

	// Delete using the position
	delResp, err := smsClient.Delete(ctx, c.Folder, position)
	if err != nil {
		return fmt.Errorf("failed to delete SMS: %w", err)
	}

	if delResp.Error != 0 {
		return fmt.Errorf("router returned error code: %d", delResp.Error)
	}

	fmt.Printf("Message with ID %d (position %d) deleted\n", msgID, position)
	return nil
}

// outputTable formats SMS messages as a markdown table
func (c *SMSCommand) outputTable(messages []model.SMSMessage) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetColWidth(180) // Set a wider column width

	if c.Folder == "inbox" {
		table.SetHeader([]string{"#", "ID", "Sender", "Message", "Date/Age"})
		for i, msg := range messages {
			table.Append([]string{
				fmt.Sprintf("%d", i+1),
				fmt.Sprintf("%d", msg.Index),
				msg.From,
				msg.Content,
				formatTime(msg.RecvTime),
			})
		}
	} else {
		table.SetHeader([]string{"#", "ID", "To", "Message", "Date/Age"})
		for i, msg := range messages {
			table.Append([]string{
				fmt.Sprintf("%d", i+1),
				fmt.Sprintf("%d", msg.Index),
				msg.To,
				msg.Content,
				formatTime(msg.SentTime),
			})
		}
	}

	table.Render()
	return nil
}

// outputJSON formats SMS messages as indented JSON
func (c *SMSCommand) outputJSON(messages []model.SMSMessage) error {
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// truncate truncates a string to max length
func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

// formatTime formats a time.Time for display
func formatTime(t interface{}) string {
	if t == nil {
		return ""
	}
	tt, ok := t.(time.Time)
	if !ok {
		return ""
	}
	// Show relative time (e.g., "2 days ago") or exact date
	now := time.Now()
	diff := now.Sub(tt)

	if diff < 0 {
		return tt.Format("02.01.2006 15:04")
	}

	hours := int(diff.Hours())
	days := hours / 24

	if days > 0 {
		return fmt.Sprintf("%d days ago", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%d hours ago", hours)
	}

	minutes := int(diff.Minutes())
	if minutes > 0 {
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	return "just now"
}
