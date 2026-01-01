# TP Link CLI

This is a basic TP Link CLI.

```bash
go install github.com/titpetric/tp-link-cli
```

Build from source:

```bash
git clone github.com:titpetric/tp-link-cli.git
cd tp-link-cli && go install .
```

It's compatible with Archer MR600. It's intended to be used for scripting jobs.

```bash
$ tp-link-cli sms
SMS Commands

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
```

As understood, the deletion mechanism for the SMS inbox is based on
order. Rather than saying which message gets deleted, you pass the
element from the list. As the element gets deleted, the order changes.

The `delete-id` command is a utility wrapping the index based delete.

## Package API

- `model/` - contains the data models related to sms commands,
- `client/` - implements `request.go` for encryption, `client.go` for API returning model types.

Acceptance tests:

- `go install .`
- `go fmt ./...`
- `go-fsck lint ./...`
- `go test -v ./...`
- `tp-link-cli sms list {folder}` - list messages on router

Folder is expected to be either "inbox" or "sent" (confirm with implementation).