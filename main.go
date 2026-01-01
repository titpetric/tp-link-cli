package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		PrintMainHelp()
		os.Exit(1)
	}

	// Check for help flags
	if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help" {
		PrintMainHelp()
		os.Exit(0)
	}

	if os.Args[1] != "sms" {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		PrintMainHelp()
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		PrintSMSHelp()
		os.Exit(1)
	}

	// Check for help on sms subcommands
	if os.Args[2] == "-h" || os.Args[2] == "--help" || os.Args[2] == "help" {
		PrintSMSHelp()
		os.Exit(0)
	}

	cmd, subcommand, err := ParseArgs(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n\n", err)
		PrintSMSHelp()
		os.Exit(1)
	}

	ctx := context.Background()

	switch subcommand {
	case "list":
		if err := cmd.ListSMS(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "read":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "error: read command requires an index\n\n")
			PrintReadHelp()
			os.Exit(1)
		}
		index, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid index: %v\n\n", err)
			PrintReadHelp()
			os.Exit(1)
		}
		if err := cmd.ReadSMS(ctx, index); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "delete":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "error: delete command requires a position\n\n")
			PrintDeleteHelp()
			os.Exit(1)
		}
		position, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid position: %v\n\n", err)
			PrintDeleteHelp()
			os.Exit(1)
		}
		if err := cmd.DeleteSMS(ctx, position); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "delete-id":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "error: delete-id command requires a message ID\n\n")
			PrintDeleteIDHelp()
			os.Exit(1)
		}
		msgID, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid id: %v\n\n", err)
			PrintDeleteIDHelp()
			os.Exit(1)
		}
		if err := cmd.DeleteSMSByID(ctx, msgID); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "send":
		// Check for help flag first
		if len(os.Args) > 3 && (os.Args[3] == "-h" || os.Args[3] == "--help" || os.Args[3] == "help") {
			PrintSendHelp()
			os.Exit(0)
		}
		if len(os.Args) < 5 {
			fmt.Fprintf(os.Stderr, "error: send command requires a phone number and message\n\n")
			PrintSendHelp()
			os.Exit(1)
		}
		number := os.Args[3]
		message := os.Args[4]
		if message == "" {
			fmt.Fprintf(os.Stderr, "error: message cannot be empty\n\n")
			PrintSendHelp()
			os.Exit(1)
		}
		if err := cmd.SendSMS(ctx, number, message); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown sms subcommand: %s\n\n", subcommand)
		PrintSMSHelp()
		os.Exit(1)
	}
}

func PrintMainHelp() {
	fmt.Fprintf(os.Stdout, `TP-Link CLI - SMS Management Tool

Usage:
  tp-link-cli [command] [options]

Commands:
  sms                 Manage SMS messages
  help, -h, --help    Show this help message

Examples:
  tp-link-cli sms list
  tp-link-cli sms read 1
  tp-link-cli sms delete 1
  tp-link-cli help

`)
}

func PrintSMSHelp() {
	fmt.Fprintf(os.Stdout, `SMS Commands

Usage:
  tp-link-cli sms <command> [options]

Commands:
  list          List all SMS messages
  read <id>     Read a specific message by ID
  delete <pos>  Delete a message by position (1-based)
  delete-id <id>    Delete a message by ID
  send <number> <message>  Send an SMS message
  help, -h, --help  Show this help message

Global Options:
  --auth=<user:pass>   Authentication credentials (default: admin:admin)
  --host=<ip>          Router IP address (default: 192.168.1.1)
  --folder=<folder>    Message folder: inbox or sent (default: inbox)
  --json               Output results as JSON

Examples:
  tp-link-cli sms list
  tp-link-cli sms list --folder=sent
  tp-link-cli sms list --json
  tp-link-cli sms read 5
  tp-link-cli sms read 5 --folder=sent
  tp-link-cli sms delete 1
  tp-link-cli sms delete-id 12345
  tp-link-cli sms send 0038612345678 "Hello, this is a test message"
  tp-link-cli sms list --host=192.168.1.100 --auth=admin:mypassword

`)
}

func PrintReadHelp() {
	fmt.Fprintf(os.Stdout, `SMS Read Command

Usage:
  tp-link-cli sms read <id> [options]

Arguments:
  <id>    The message ID to read

Options:
  --auth=<user:pass>   Authentication credentials (default: admin:admin)
  --host=<ip>          Router IP address (default: 192.168.1.1)
  --folder=<folder>    Message folder: inbox or sent (default: inbox)
  --json               Output result as JSON

Examples:
  tp-link-cli sms read 1
  tp-link-cli sms read 5 --folder=sent
  tp-link-cli sms read 3 --json

`)
}

func PrintDeleteHelp() {
	fmt.Fprintf(os.Stdout, `SMS Delete Command (by Position)

Usage:
  tp-link-cli sms delete <position> [options]

Arguments:
  <position>    The 1-based position of the message to delete (as shown in list output)

Options:
  --auth=<user:pass>   Authentication credentials (default: admin:admin)
  --host=<ip>          Router IP address (default: 192.168.1.1)
  --folder=<folder>    Message folder: inbox or sent (default: inbox)

Note: Use 'delete-id' to delete by message ID instead of position.

Examples:
  tp-link-cli sms delete 1
  tp-link-cli sms delete 5 --folder=sent

`)
}

func PrintDeleteIDHelp() {
	fmt.Fprintf(os.Stdout, `SMS Delete Command (by ID)

Usage:
  tp-link-cli sms delete-id <id> [options]

Arguments:
  <id>    The message ID to delete (as shown in the ID column)

Options:
  --auth=<user:pass>   Authentication credentials (default: admin:admin)
  --host=<ip>          Router IP address (default: 192.168.1.1)
  --folder=<folder>    Message folder: inbox or sent (default: inbox)

Note: Use 'delete' to delete by position instead of ID.

Examples:
  tp-link-cli sms delete-id 12345
  tp-link-cli sms delete-id 67890 --folder=sent

`)
}

func PrintSendHelp() {
	fmt.Fprintf(os.Stdout, `SMS Send Command

Usage:
  tp-link-cli sms send <number> <message> [options]

Arguments:
  <number>    The phone number to send to
  <message>   The message text to send

Options:
  --auth=<user:pass>   Authentication credentials (default: admin:admin)
  --host=<ip>          Router IP address (default: 192.168.1.1)

Examples:
  tp-link-cli sms send 0038612345678 "Hello, this is a test message"
  tp-link-cli sms send 0038612345678 "Test message" --auth=admin:mypassword
  tp-link-cli sms send +41791234567 "Important notification"

`)
}
