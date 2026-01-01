# TP Link CLI

This is a basic TP Link CLI.

```
git clone github.com:titpetric/tp-link-cli.git
cd tp-link-cli && go install .
```

It's compatible with Archer MR600. It's intended to be used for scripting jobs.
It allows a program or human to do the following:

- `tp-link-cli sms list` - a list of the SMS inbox,
- `tp-link-cli sms read N` - read an SMS,
- `tp-link-cli sms delete N` - delete a SMS.

As understood, the deletion mechanism for the SMS inbox is based on
order. Rather than saying which message gets deleted, you pass the
element from the list. As the element gets deleted, the order changes.

Example, sms list:

```
| # | Sender | Message | Date/Age |
|---|--------|---------|----------|
| 1 | GOVT   | New educational programs for 2026. |  2 days ago |
| 2 | BANKS  | Your standing orders have been deducted. | 15.12.2025 10:00 |
```

It's intended to use the `tablewriter` or `tabwriter` package for
printing the markdown table (third party).

The sms command should take some parameters:

```
--auth=admin:default   The user/password pair for your router.
--host=192.168.1.1     The host of your router.
--json                 Render pretty printed json instead of markdown.
```

## Package API

- `model/` - contains the data models related to sms commands,
- `client/` - implements `request.go` for encryption, `client.go` for API returning model types.

The expected client API is:

- type SMSClient struct {...}
- func NewSMSClient(*Options (auth, host)) (*SMSClient)
- func (s *SMSClient) List(context.Context, folder string) (*model.ListResponse, error)
- func (s *SMSClient) Read(context.Context, folder string, index int) (*model.ReadResponse, error)
- func (s *SMSClient) Delete(content.Context, folder string, index int) (*model.DeleteResponse, error)
- func (s *SMSClient) Send(context.Context, number string, message string) (*model.SendResponse, error)

Acceptance tests:

- `go install .`
- `go fmt ./...`
- `go-fsck lint ./...`
- `tp-link-cli sms list {folder}` - list messages on router

Folder is expected to be either "inbox" or "sent" (confirm with implementation).