package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: tp-link-cli sms <command> [options]\n")
		os.Exit(1)
	}

	if os.Args[1] != "sms" {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}

	cmd, subcommand, err := ParseArgs(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	switch subcommand {
	case "list":
		if err := cmd.ListSMS(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown sms subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}
