# TP-Link CLI - Work Progress

## Objective
Get a 0 response (no error) from the router login sequence and obtain a valid cookie for making SMS inbox read requests.

## Things I've tried

1. **Analyzed existing implementations**
   - Reviewed Python implementation in src2/ (archer.py, MyAESCrypto.py, MyRSACrypto.py)
   - Reviewed Go implementation in client/ (client.go, encryption.go, protocol.go)
   - Both implementations perform AES-CBC encryption with RSA signatures

2. **Created Python request dumping script (src2/dump_request_details.py)**
   - Captures complete request/response lifecycle
   - Logs AES encryption keys, URLs, query parameters, request bodies
   - Outputs encryption parameters (nn, ee, seq)
   - Dumps session cookies
   - Saves to request_dump.json for detailed analysis

3. **Created Go integration tests (client/integration_test.go)**
   - Added `// +build integration` tag for test separation
   - Tests target 192.168.1.1 with admin/default credentials
   - Coverage: Login, ListSMS, ReadSMS, FullWorkflow, CookieHandling
   - Run with: `go test -v -tags=integration ./client`

4. **Verified Go encryption implementation**
   - Ran `go test -v ./client -run TestAESEncryption`
   - All encryption tests PASS
   - Verified test vectors match Python implementation
   - AES-CBC encryption and decryption working correctly
   - RSA signature generation working

5. **Code structure analysis**
   - Go client follows these steps:
     1. GET / (initial page)
     2. POST /cgi/getParm (fetch RSA params: nn, ee, seq)
     3. GET /img/loading.gif
     4. POST /cgi/getBusy
     5. POST /cgi/login with encrypted credentials
     6. Extract session cookie
     7. POST /cgi_gdpr with encrypted SMS requests

## Things to try next

1. **CRITICAL: Investigate error code 71234**
   - Go login returns: `$.ret=71234;` (authentication failed)
   - Error is consistent across multiple attempts
   - Suggests signature or AES decryption validation failure
   - Need to determine: signature format or AES encryption mismatch

2. **Compare AES encryption implementations**
   - Go: Uses crypto/aes with PKCS7 padding
   - Python: Uses custom MyAESCrypto with specific format
   - Test vector: Both should encrypt "admin\ndefault" identically
   - Known issue: Python MyAESCrypto has bugs (genKey returns None)

3. **Verify signature generation matches exactly**
   - Signature format for login: `key=...&iv=...&h=...&s=...`
   - Go constructs: aesKeyStr + "&h=" + hash + "&s=" + (seq+dataLen)
   - Python constructs: same format in archer.py line 85
   - Check if RSA encryption of signature string is identical

4. **Debug AES key generation**
   - Both use: time.time() * 1e6 + random(0-999)
   - Python: returns integer, must be converted to string
   - Go: generates timestamp in microseconds
   - Test: use fixed timestamps to compare directly

5. **Once login works (error code 0)**
   - Verify JSESSIONID cookie is returned
   - Test TokenID extraction from homepage
   - Test SMS list/read operations
   - Verify cookie reuse across requests

## Key Files Created/Modified

- `/client/integration_test.go` - Integration tests with build tag
- `/src2/dump_request_details.py` - Request/response dumper
- `/WORK.md` - This file (progress tracking)

## Status Summary

**Completed:**
- ✓ Python dump script created (src2/dump_request_details.py)
- ✓ Go integration tests created (client/integration_test.go)
- ✓ Go encryption implementation verified (tests passing)
- ✓ Code analysis and documentation

**Next Steps:**
- Run Python dump script against real router (requires router access)
- Run Go integration tests against real router (requires router access)
- Compare request/response formats and identify differences
- Fix any identified issues
- Verify SMS operations work with valid cookies

**Key Test Commands:**
```bash
# Encryption verification (should pass)
go test -v ./client -run TestAESEncryption

# Integration tests (requires 192.168.1.1 with admin/default)
go test -v -tags=integration ./client

# Specific integration test
go test -v -tags=integration -run TestIntegration_Login ./client

# Python request dump (requires real router)
cd src2 && python3 dump_request_details.py
```
