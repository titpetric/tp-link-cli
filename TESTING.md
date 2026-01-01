# Testing Guide - TP-Link CLI

## Overview

This document describes how to use the testing and debugging tools for the TP-Link SMS CLI.

## Unit Tests (No Router Required)

### Run all encryption tests
```bash
go test -v ./client -run "TestAES|TestEncryption|TestProtocol"
```

### Run specific encryption test
```bash
go test -v ./client -run TestAESEncryptionVectors
```

Expected output: All tests PASS

### Run protocol tests
```bash
go test -v ./client -run TestProtocol
```

## Integration Tests (Requires Router at 192.168.1.1)

**Prerequisites:**
- TP-Link Archer MR600 router (or compatible)
- Router IP: 192.168.1.1
- Username: admin
- Password: default

### Build integration test binary
```bash
go test -v -tags=integration ./client -c
```

This creates `./client/client.test` executable.

### Run all integration tests
```bash
go test -v -tags=integration ./client
```

### Run specific integration test
```bash
# Test login
go test -v -tags=integration -run TestIntegration_Login ./client

# Test SMS list
go test -v -tags=integration -run TestIntegration_ListSMS ./client

# Test SMS read
go test -v -tags=integration -run TestIntegration_ReadSMS ./client

# Test full workflow
go test -v -tags=integration -run TestIntegration_FullWorkflow ./client

# Test encryption parameters
go test -v -tags=integration -run TestIntegration_EncryptionParameters ./client

# Test cookie handling
go test -v -tags=integration -run TestIntegration_CookieHandling ./client
```

## Python Request Dumper (Requires Router at 192.168.1.1)

### Run the request dumper
```bash
cd src2
python3 dump_request_details.py
```

### What it outputs

The script logs to stdout and saves detailed JSON to `request_dump.json`:

**Stdout output shows:**
- Connection steps (login, encryption param retrieval, etc.)
- Encryption parameters (nn, ee, seq from router)
- AES key and IV (derived from current timestamp)
- Login request details
- Session cookies obtained

**request_dump.json contains:**
- Metadata (timestamp, router, credentials)
- All requests made with method, URL, headers, body
- Encryption parameters used
- AES keys generated
- Session cookies obtained

### Use case

Compare Python implementation with Go implementation:
1. Run Python dumper: `python3 dump_request_details.py`
2. Save output: `cp request_dump.json request_dump_python.json`
3. Analyze what requests/responses occur
4. Run Go integration test with same credentials
5. Compare debug output from both implementations

## Debugging Login Issues

### Step 1: Verify encryption works
```bash
go test -v ./client -run TestAESEncryptionVectors
```
Must PASS.

### Step 2: Check if router is reachable
```bash
ping 192.168.1.1
curl -I http://192.168.1.1/
```

### Step 3: Run Python version first
```bash
cd src2
python3 dump_request_details.py
```
Look for successful login (Session ID obtained).

### Step 4: Run Go version
```bash
go test -v -tags=integration -run TestIntegration_Login ./client 2>&1
```
Compare debug output with Python version.

### Step 5: Look for differences in:
- AES key generation (numeric string format)
- URL encoding of encrypted data
- Request headers
- Cookie handling
- Response parsing

## Expected Behavior

### Successful Login (Error Code 0)
```
✓ Login successful!
✓ Session ID: <JSESSIONID value>
✓ Token ID: <token value>
```

### Successful SMS List (Error Code 0)
```
✓ SMS List successful!
✓ Message count: <N>
✓ Error code: 0 (0 = no error)
```

### Successful SMS Read
```
✓ SMS Read successful!
✓ Error code: 0 (0 = no error)
  From: <sender>
  Content: <message>
  Received: <timestamp>
```

## Troubleshooting

### Test fails with "connection refused"
- Router is offline
- Wrong IP address
- Firewall blocking access

### Test fails with "authentication failed"
- Wrong credentials
- Router expects different username/password
- Look at Python dumper output for comparison

### Test fails with "invalid cookie"
- Cookie extraction parsing is wrong
- Look at JSESSIONID in response headers

### Test fails with non-zero error code
- Encryption parameters incorrect
- Request format differs from router expectations
- AES key generation differs from Python

## Comparing Go vs Python

```bash
# Terminal 1: Run Python dumper
cd src2 && python3 dump_request_details.py > python_output.txt 2>&1

# Terminal 2: Run Go integration test
go test -v -tags=integration -run TestIntegration_Login ./client > go_output.txt 2>&1

# Compare outputs
diff python_output.txt go_output.txt

# Or view both
cat python_output.txt
cat go_output.txt
```

Look for differences in:
1. AES key string format
2. Encrypted data (base64 output)
3. RSA signature
4. URL construction
5. Request headers
6. Cookie extraction
