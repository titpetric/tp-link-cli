package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"tp-link-cli/client"
	"tp-link-cli/model"

	"github.com/olekukonko/tablewriter"
)

// SMSCommand handles SMS operations
type SMSCommand struct {
	Auth   string
	Host   string
	JSON   bool
	Folder string
}

// ParseArgs parses command-line arguments
func ParseArgs(args []string) (*SMSCommand, string, error) {
	if len(args) == 0 {
		return nil, "", fmt.Errorf("usage: tp-link-cli sms <command> [options]")
	}

	cmd := &SMSCommand{
		Auth: "admin:default",
		Host: "192.168.1.1",
	}

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
	smsClient, err := client.NewSMSClient(&client.Options{
		Auth: c.Auth,
		Host: c.Host,
	})
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

// outputTable formats SMS messages as a markdown table
func (c *SMSCommand) outputTable(messages []model.SMSMessage) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)

	if c.Folder == "inbox" {
		table.SetHeader([]string{"#", "Sender", "Message", "Date/Age"})
		for i, msg := range messages {
			table.Append([]string{
				fmt.Sprintf("%d", i+1),
				msg.From,
				truncate(msg.Content, 50),
				formatTime(msg.RecvTime),
			})
		}
	} else {
		table.SetHeader([]string{"#", "To", "Message", "Date/Age"})
		for i, msg := range messages {
			table.Append([]string{
				fmt.Sprintf("%d", i+1),
				msg.To,
				truncate(msg.Content, 50),
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
