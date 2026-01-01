package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"tp-link-cli/model"
)

// Options holds client configuration
type Options struct {
	Auth string // "username:password"
	Host string // "192.168.1.1" or "http://192.168.1.1"
}

// SMSClient communicates with TP-Link router
type SMSClient struct {
	baseURL    string
	username   string
	password   string
	sessionID  string
	tokenID    string
	enc        *Encryption
	proto      *Protocol
	httpClient *http.Client
}

// NewSMSClient creates a new SMS client
func NewSMSClient(opts *Options) (*SMSClient, error) {
	// Parse auth
	var username, password string
	parts := strings.Split(opts.Auth, ":")
	if len(parts) == 2 {
		username = parts[0]
		password = parts[1]
	} else {
		// If no colon, use the same string for both username and password
		username = opts.Auth
		password = opts.Auth
	}

	// Parse host
	baseURL := opts.Host
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}

	jar, _ := cookiejar.New(&cookiejar.Options{})
	return &SMSClient{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		enc:        NewEncryption(),
		proto:      NewProtocol(),
		httpClient: &http.Client{Jar: jar},
	}, nil
}

// connect performs authentication and setup
func (c *SMSClient) connect(ctx context.Context) error {
	// Step 0: Fetch initial page to establish cookies
	initReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return err
	}
	initReq.Header.Set("Accept", "*/*")
	initReq.Header.Set("Accept-Encoding", "gzip, deflate")
	initReq.Header.Set("Accept-Language", "en-US,en;q=0.9")
	initReq.Header.Set("Cache-Control", "no-cache")
	initReq.Header.Set("Connection", "keep-alive")
	initReq.Header.Set("Pragma", "no-cache")
	initReq.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
	initReq.Header.Set("Referer", c.baseURL)

	initResp, err := c.httpClient.Do(initReq)
	if err != nil {
		return fmt.Errorf("failed to fetch homepage: %w", err)
	}
	defer initResp.Body.Close()
	if initResp.StatusCode != 200 {
		body, _ := io.ReadAll(initResp.Body)
		return fmt.Errorf("homepage returned status %d: %s", initResp.StatusCode, string(body))
	}

	// Step 1: Fetch encryption parameters
	parmURL := c.baseURL + "/cgi/getParm"
	req, err := http.NewRequestWithContext(ctx, "POST", parmURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
	req.Header.Set("Referer", c.baseURL)
	req.Header.Set("Host", strings.TrimPrefix(c.baseURL, "http://"))

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get encryption params: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	// Step 2: Parse encryption parameters
	ee, nn, seq, err := ParseEncryptionParams(string(body))
	if err != nil {
		return fmt.Errorf("failed to parse encryption params: %w", err)
	}

	// Step 3: Configure encryption
	if err := c.enc.SetRSAKey(nn, ee); err != nil {
		return err
	}
	c.enc.GenAESKey()
	// Convert seq string to int
	seqNum := 0
	fmt.Sscanf(seq, "%d", &seqNum)
	c.enc.SetSeq(seqNum)
	// Set hash based on username and password
	c.enc.SetHash(c.username, c.password)

	// Step 3b: Load loading.gif (mimics browser behavior)
	// NOTE: This also updates the default Accept header used for subsequent requests
	gifURL := c.baseURL + "/img/loading.gif"
	gifReq, err := http.NewRequestWithContext(ctx, "GET", gifURL, nil)
	if err == nil {
		gifReq.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
		gifReq.Header.Set("Accept-Encoding", "gzip, deflate")
		gifReq.Header.Set("Accept-Language", "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7")
		gifReq.Header.Set("Cache-Control", "no-cache")
		gifReq.Header.Set("Connection", "keep-alive")
		gifReq.Header.Set("Pragma", "no-cache")
		gifReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.67 Safari/537.36")
		gifReq.Header.Set("Referer", c.baseURL)
		gifReq.Header.Set("Host", strings.TrimPrefix(c.baseURL, "http://"))
		if resp, err := c.httpClient.Do(gifReq); err == nil {
			resp.Body.Close()
		}
	}

	// Step 3c: Check if router is busy
	busyURL := c.baseURL + "/cgi/getBusy"
	busyReq, err := http.NewRequestWithContext(ctx, "POST", busyURL, nil)
	if err == nil {
		busyReq.Header.Set("Referer", c.baseURL)
		if resp, err := c.httpClient.Do(busyReq); err == nil {
			resp.Body.Close()
		}
	}

	// Step 4: Authenticate
	authData := c.username + "\n" + c.password
	encrypted := c.enc.AESEncrypt(authData, true)

	// Debug output
	fmt.Fprintf(os.Stderr, "DEBUG: Hash: %s\n", c.enc.hash)
	fmt.Fprintf(os.Stderr, "DEBUG: Seq: %d\n", c.enc.seq)
	fmt.Fprintf(os.Stderr, "DEBUG: AES Key String: %s\n", c.enc.aesKeyStr)
	fmt.Fprintf(os.Stderr, "DEBUG: Encrypted Data (first 50 chars): %.50s...\n", encrypted.Data)
	fmt.Fprintf(os.Stderr, "DEBUG: Encrypted Sign (first 50 chars): %.50s...\n", encrypted.Sign)

	// URL encode the data - replace specific characters as per Python code
	// data.replace('=', '%3D').replace('+', '%2B')
	encodedData := strings.ReplaceAll(encrypted.Data, "=", "%3D")
	encodedData = strings.ReplaceAll(encodedData, "+", "%2B")

	fmt.Fprintf(os.Stderr, "DEBUG: Encoded Data: %s\n", encodedData)
	fmt.Fprintf(os.Stderr, "DEBUG: Encrypted.Data was: %s\n", encrypted.Data)

	loginURL := fmt.Sprintf("%s/cgi/login?data=%s&sign=%s&Action=1&LoginStatus=0",
		c.baseURL, encodedData, encrypted.Sign)
	
	fmt.Fprintf(os.Stderr, "GO LOGIN REQUEST\n")
	fmt.Fprintf(os.Stderr, "URL: %s\n", loginURL)
	fmt.Fprintf(os.Stderr, "Encrypted data (plain): %s\n", encrypted.Data)
	fmt.Fprintf(os.Stderr, "Encrypted data (encoded): %s\n", encodedData)
	fmt.Fprintf(os.Stderr, "Signature: %s\n", encrypted.Sign)

	authReq, err := http.NewRequestWithContext(ctx, "POST", loginURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}
	authReq.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
	authReq.Header.Set("Accept-Encoding", "gzip, deflate")
	authReq.Header.Set("Accept-Language", "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7")
	authReq.Header.Set("Cache-Control", "no-cache")
	authReq.Header.Set("Connection", "keep-alive")
	authReq.Header.Set("Pragma", "no-cache")
	authReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.67 Safari/537.36")
	authReq.Header.Set("Referer", c.baseURL)
	authReq.Header.Set("Host", strings.TrimPrefix(c.baseURL, "http://"))
	authReq.Header.Set("Cookie", "loginErrorShow=1")
	authReq.Header.Set("Origin", c.baseURL)

	fmt.Fprintf(os.Stderr, "\nHeaders:\n")
	for k, v := range authReq.Header {
		fmt.Fprintf(os.Stderr, "  %s: %s\n", k, v[0])
	}

	authResp, err := c.httpClient.Do(authReq)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer authResp.Body.Close()
	
	// Read response body
	authRespBody, err := io.ReadAll(authResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}
	
	fmt.Fprintf(os.Stderr, "\nGO LOGIN RESPONSE\n")
	fmt.Fprintf(os.Stderr, "Status: %d\n", authResp.StatusCode)
	fmt.Fprintf(os.Stderr, "Body: %s\n", string(authRespBody))
	fmt.Fprintf(os.Stderr, "Response Headers:\n")
	for k, v := range authResp.Header {
		fmt.Fprintf(os.Stderr, "  %s: %s\n", k, v[0])
	}

	// Step 5: Extract session ID from Set-Cookie
	for _, cookie := range authResp.Cookies() {
		fmt.Fprintf(os.Stderr, "DEBUG: Found cookie: %s=%s\n", cookie.Name, cookie.Value)
		if cookie.Name == "JSESSIONID" {
			c.sessionID = cookie.Value
			break
		}
	}
	
	// Also check jar for cookies (cookiejar auto-populates)
	if c.sessionID == "" && c.httpClient.Jar != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: Checking cookie jar...\n")
		jarURL, _ := url.Parse(c.baseURL)
		jarCookies := c.httpClient.Jar.Cookies(jarURL)
		fmt.Fprintf(os.Stderr, "DEBUG: Jar cookies: %v\n", jarCookies)
		for _, cookie := range jarCookies {
			fmt.Fprintf(os.Stderr, "DEBUG: Jar cookie: %s=%s\n", cookie.Name, cookie.Value)
			if cookie.Name == "JSESSIONID" {
				c.sessionID = cookie.Value
				break
			}
		}
	}
	
	// Log auth response for debugging
	fmt.Fprintf(os.Stderr, "DEBUG: Session ID: %s\n", c.sessionID)
	fmt.Fprintf(os.Stderr, "DEBUG: All response cookies: %v\n", authResp.Cookies())

	if c.sessionID == "" {
		// Try to read response body for error details
		return fmt.Errorf("failed to obtain session ID\nStatus: %d\nResponse: %s\nAuth Req URL: %s", authResp.StatusCode, string(authRespBody), loginURL)
	}

	// Step 6: Fetch token ID
	homeURL := c.baseURL + "/"
	homeReq, err := http.NewRequestWithContext(ctx, "GET", homeURL, nil)
	if err != nil {
		return err
	}
	homeReq.Header.Set("Referer", c.baseURL)
	homeReq.Header.Set("Cookie", "loginErrorShow=1; JSESSIONID="+c.sessionID)

	homeResp, err := c.httpClient.Do(homeReq)
	if err != nil {
		return fmt.Errorf("failed to fetch homepage: %w", err)
	}
	defer homeResp.Body.Close()

	homeBody, err := io.ReadAll(homeResp.Body)
	if err != nil {
		return err
	}

	tokenRegex := regexp.MustCompile(`(?i)var\s+token\s*=\s*"([a-f0-9]+)"`)
	matches := tokenRegex.FindStringSubmatch(string(homeBody))
	if len(matches) < 2 {
		// Log the response for debugging
		bodyStr := string(homeBody)
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500]
		}
		return fmt.Errorf("failed to extract token ID from homepage\nResponse: %s", bodyStr)
	}
	c.tokenID = matches[1]

	return nil
}

// execute sends an encrypted request to the router
func (c *SMSClient) execute(ctx context.Context, reqs []Request) (Response, error) {
	// Ensure we're authenticated
	if c.tokenID == "" {
		if err := c.connect(ctx); err != nil {
			return Response{}, err
		}
	}

	// Build and encrypt frame
	dataFrame := c.proto.MakeDataFrame(reqs)
	encrypted := c.enc.AESEncrypt(dataFrame, false)
	payload := fmt.Sprintf("sign=%s\r\ndata=%s\r\n", encrypted.Sign, encrypted.Data)

	cgiURL := c.baseURL + "/cgi_gdpr"
	req, err := http.NewRequestWithContext(ctx, "POST", cgiURL, strings.NewReader(payload))
	if err != nil {
		return Response{}, err
	}

	req.Header.Set("Referer", c.baseURL)
	req.Header.Set("Cookie", "loginErrorShow=1; JSESSIONID="+c.sessionID)
	req.Header.Set("TokenID", c.tokenID)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	// Decrypt response
	decrypted, err := c.enc.AESDecrypt(string(respBody))
	if err != nil {
		return Response{}, fmt.Errorf("failed to decrypt response: %w", err)
	}

	// Parse response
	parsed := c.proto.FromDataFrame(decrypted)
	return c.proto.PrettifyResponse(parsed), nil
}

// List retrieves SMS messages from the specified folder
func (c *SMSClient) List(ctx context.Context, folder string) (*model.ListResponse, error) {
	if folder == "" {
		folder = "inbox"
	}

	var reqs []Request

	// Reset cursor
	reqs = append(reqs, Request{
		Method:     ActSet,
		Controller: "LTE_SMS_RECVMSGBOX",
		Attrs: map[string]interface{}{
			"PageNumber": 1,
		},
	})

	// Get messages
	var controller string
	var attrs []string
	if folder == "inbox" {
		controller = "LTE_SMS_RECVMSGENTRY"
		attrs = []string{"index", "from", "content", "receivedTime", "unread"}
	} else if folder == "sent" {
		controller = "LTE_SMS_SENDMSGENTRY"
		attrs = []string{"index", "to", "content", "sendTime"}
	} else {
		return nil, fmt.Errorf("invalid folder: %s", folder)
	}

	reqs = append(reqs, Request{
		Method:     ActGL,
		Controller: controller,
		Attrs:      attrs,
	})

	resp, err := c.execute(ctx, reqs)
	if err != nil {
		return nil, err
	}

	// Convert response to model
	result := &model.ListResponse{
		Error: resp.Error,
	}

	for _, obj := range resp.Data {
		msg := c.rawToSMSMessage(obj, folder)
		result.Data = append(result.Data, msg)
	}

	return result, nil
}

// Read retrieves a specific SMS message
func (c *SMSClient) Read(ctx context.Context, folder string, index int) (*model.ReadResponse, error) {
	if folder == "" {
		folder = "inbox"
	}

	var controller string
	if folder == "inbox" {
		controller = "LTE_SMS_RECVMSGENTRY"
	} else if folder == "sent" {
		controller = "LTE_SMS_SENDMSGENTRY"
	} else {
		return nil, fmt.Errorf("invalid folder: %s", folder)
	}

	reqs := []Request{
		{
			Method:     ActGet,
			Controller: controller,
			Stack:      fmt.Sprintf("%d,0,0,0,0,0", index),
			Attrs:      []string{"index", "from", "content", "receivedTime", "unread"},
		},
	}

	resp, err := c.execute(ctx, reqs)
	if err != nil {
		return nil, err
	}

	result := &model.ReadResponse{
		Error: resp.Error,
	}

	for _, obj := range resp.Data {
		msg := c.rawToSMSMessage(obj, folder)
		result.Data = append(result.Data, msg)
	}

	return result, nil
}

// Delete removes an SMS message
func (c *SMSClient) Delete(ctx context.Context, folder string, index int) (*model.DeleteResponse, error) {
	if folder == "" {
		folder = "inbox"
	}

	var controller string
	if folder == "inbox" {
		controller = "LTE_SMS_RECVMSGENTRY"
	} else if folder == "sent" {
		controller = "LTE_SMS_SENDMSGENTRY"
	} else {
		return nil, fmt.Errorf("invalid folder: %s", folder)
	}

	reqs := []Request{
		{
			Method:     ActDel,
			Controller: controller,
			Stack:      fmt.Sprintf("%d,0,0,0,0,0", index),
		},
	}

	resp, err := c.execute(ctx, reqs)
	if err != nil {
		return nil, err
	}

	result := &model.DeleteResponse{
		Error: resp.Error,
	}

	for _, obj := range resp.Data {
		msg := c.rawToSMSMessage(obj, folder)
		result.Data = append(result.Data, msg)
	}

	return result, nil
}

// Send sends an SMS message
func (c *SMSClient) Send(ctx context.Context, number, message string) (*model.SendResponse, error) {
	reqs := []Request{
		{
			Method:     ActSet,
			Controller: "LTE_SMS_SENDNEWMSG",
			Attrs: map[string]interface{}{
				"index":       1,
				"to":          number,
				"textContent": message,
			},
		},
	}

	resp, err := c.execute(ctx, reqs)
	if err != nil {
		return nil, err
	}

	result := &model.SendResponse{
		Error: resp.Error,
	}

	for _, obj := range resp.Data {
		msg := c.rawToSMSMessage(obj, "sent")
		result.Data = append(result.Data, msg)
	}

	return result, nil
}

// rawToSMSMessage converts raw response data to SMSMessage
func (c *SMSClient) rawToSMSMessage(obj map[string]interface{}, folder string) model.SMSMessage {
	msg := model.SMSMessage{}

	if v, ok := obj["index"]; ok {
		if idx, ok := v.(int); ok {
			msg.Index = idx
		}
	}
	if v, ok := obj["from"]; ok {
		msg.From = v.(string)
	}
	if v, ok := obj["to"]; ok {
		msg.To = v.(string)
	}
	if v, ok := obj["content"]; ok {
		msg.Content = v.(string)
	}
	if v, ok := obj["receivedTime"]; ok {
		msg.RecvTime = v.(time.Time)
	}
	if v, ok := obj["sendTime"]; ok {
		msg.SentTime = v.(time.Time)
	}
	if v, ok := obj["unread"]; ok {
		msg.Unread = v.(bool)
	}

	return msg
}
