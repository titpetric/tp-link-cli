# TP Link CLI

This is a basic TP Link CLI. It provides an API over the routers SMS features.

It's tested/compatible with Archer MR600.

```bash
go install github.com/titpetric/tp-link-cli@latest
```

Build from source:

```bash
git clone github.com:titpetric/tp-link-cli.git
cd tp-link-cli && go install .
```

The tool is intended to be used for automation jobs. I'm automating a
challenge/response system based on ISP restrictions, but the CLI can be
used to implement an SMS gateway as well.

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

You can provide `TP_LINK_CLI_HOST` and `TP_LINK_CLI_AUTH` as environment
variables, avoiding the need to pass `--host` or `--auth` args.

As implemented, the deletion mechanism for the SMS inbox is based on
order. Rather than saying which message gets deleted, you pass the
element from the list. As the element gets deleted, the order changes.
The deletion is based on the session, so if you want to delete
everything but the last 3 messages, you'd loop over `sms delete 4 ; sms
list ; ...`.

The `delete-id` command is a utility wrapping the index based delete.

## Package API

- `model/` - contains the data models related to sms commands,
- `client/` - implements `request.go` for encryption, `client.go` for API returning model types.

Acceptance tests (see Taskfile):

- `go install .`
- `go fmt ./...`
- `go-fsck lint ./...`
- `go test -v ./...`
- `tp-link-cli sms list {folder}` - list messages on router

Folder is expected to be either "inbox" or "sent" (confirm with implementation).

## Unimplemented

The tool doesn't implement folder pagination. The device returns the
latest 8 messages available for inbox/sent. It can be extended if you
need this, say to provide a backup of the messages.

## License

Public domain.
