# Quick Start - TP-Link CLI Testing

## One-Minute Overview

You have two implementations (Python and Go) for communicating with a TP-Link router:
- **Python** (src2/): Reference implementation, works against real routers
- **Go** (client/): CLI version, needs debugging to match router expectations

Goal: Get a 0 error response from login and obtain valid cookies for SMS operations.

## Check Everything Works (No Router Needed)

```bash
# Test Go encryption
go test -v ./client -run TestAESEncryption

# Should see: PASS
```

## Test Against Real Router (192.168.1.1 required)

### Option 1: Run Python Dumper (Reference Implementation)
```bash
cd src2
python3 dump_request_details.py
# Output: Logs all requests/responses + JSON dump
```

### Option 2: Run Go Integration Tests
```bash
go test -v -tags=integration -run TestIntegration_Login ./client
# Should eventually match Python's behavior
```

### Option 3: Compare Both
```bash
# Terminal 1
cd src2 && python3 dump_request_details.py 2>&1 | tee python.log

# Terminal 2  
go test -v -tags=integration -run TestIntegration_Login ./client 2>&1 | tee golang.log

# Then diff to find differences
diff python.log golang.log
```

## What to Look For

**Success (error code 0):**
```
✓ Login successful!
✓ Session ID: XXXXXX
✓ Token ID: XXXXXX
✓ SMS List successful!
✓ Error code: 0
```

**Failure (non-zero error code):**
```
⚠ Router returned error code: 1
```

## Troubleshooting

1. **No router?** Skip integration tests, just run unit tests
2. **Router not responding?** Check IP address and ping first
3. **Wrong credentials?** Verify admin/default are correct
4. **Different error codes?** Compare Python vs Go request format

## Files You Created

| File | Purpose |
|------|---------|
| `WORK.md` | Progress tracking |
| `TESTING.md` | Detailed testing guide |
| `src2/dump_request_details.py` | Python request dumper |
| `client/integration_test.go` | Go integration tests |

## Key Commands

```bash
# Unit tests (always works)
go test -v ./client -run TestAES

# Integration tests (needs router at 192.168.1.1:admin/default)
go test -v -tags=integration ./client

# Python equivalent (needs router at 192.168.1.1:admin/default)
cd src2 && python3 dump_request_details.py

# Specific test
go test -v -tags=integration -run TestIntegration_Login ./client
```

## Next Steps

1. Access a real TP-Link Archer MR600 router
2. Run: `cd src2 && python3 dump_request_details.py`
3. Compare output with Go tests
4. Identify and fix any differences
5. Verify SMS list/read operations work

See `TESTING.md` for detailed procedures.
