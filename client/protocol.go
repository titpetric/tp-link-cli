package client

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	ActGet = 1
	ActSet = 2
	ActDel = 4
	ActGL  = 5
	ActGS  = 6
	ActCGI = 8
)

// Request represents a router protocol request
type Request struct {
	Method     int
	Controller string
	Stack      string
	Attrs      interface{} // map[string]interface{} or []string
}

// Response represents a router protocol response
type Response struct {
	Error int
	Data  []map[string]interface{}
}

// Protocol handles encoding and decoding of TP-Link router protocol messages
type Protocol struct{}

// NewProtocol creates a new Protocol handler
func NewProtocol() *Protocol {
	return &Protocol{}
}

// MakeDataFrame creates a protocol data frame from a request
func (p *Protocol) MakeDataFrame(requests []Request) string {
	var sections []map[string]interface{}

	for _, req := range requests {
		stack := req.Stack
		if stack == "" {
			stack = "0,0,0,0,0,0"
		}
		pStack := "0,0,0,0,0,0"

		attrs := p.toKV(req.Attrs)
		nbAttrs := strings.Count(attrs, "\r\n")

		sections = append(sections, map[string]interface{}{
			"method":     req.Method,
			"controller": req.Controller,
			"stack":      stack,
			"pStack":     pStack,
			"attrs":      attrs,
			"nbAttrs":    nbAttrs,
		})
	}

	// Build header
	var methods []string
	for _, sec := range sections {
		methods = append(methods, fmt.Sprintf("%d", sec["method"]))
	}
	header := strings.Join(methods, "&")

	// Build data
	var dataStr string
	for i, s := range sections {
		dataStr += fmt.Sprintf("[%s#%s#%s]%d,%d\r\n%s",
			s["controller"], s["stack"], s["pStack"], i, s["nbAttrs"], s["attrs"])
	}

	return header + "\r\n" + dataStr
}

// toKV converts an object to key-value format
func (p *Protocol) toKV(data interface{}) string {
	switch v := data.(type) {
	case string:
		return v
	case []string:
		return strings.Join(v, "\r\n") + "\r\n"
	case map[string]interface{}:
		var result string
		for key, val := range v {
			valStr := ""
			if val != nil {
				switch vv := val.(type) {
				case string:
					vv = strings.ReplaceAll(vv, "\n", "\u0012")
					vv = strings.ReplaceAll(vv, "\r", "\u0012")
					valStr = vv
				case int:
					valStr = fmt.Sprintf("%d", vv)
				default:
					valStr = fmt.Sprintf("%v", vv)
				}
				result += fmt.Sprintf("%s=%s\r\n", key, valStr)
			} else {
				result += fmt.Sprintf("%s\r\n", key)
			}
		}
		return result
	}
	return ""
}

// FromDataFrame parses a protocol data frame response
func (p *Protocol) FromDataFrame(frame string) Response {
	lines := strings.Split(strings.TrimSpace(frame), "\n")
	var data []map[string]interface{}
	var currentObject map[string]interface{}
	var errorCode int

	objectHeaderRegex := regexp.MustCompile(`^\[\d,\d,\d,\d,\d,\d\]\d`)
	objectAttrRegex := regexp.MustCompile(`^([a-zA-Z0-9]+)=(.*)$`)
	frameErrorRegex := regexp.MustCompile(`^\[error\](\d+)$`)

	for _, line := range lines {
		// Check for object header
		if objectHeaderRegex.MatchString(line) {
			if currentObject != nil {
				data = append(data, currentObject)
			}
			currentObject = make(map[string]interface{})
			continue
		}

		// Check for error code
		match := frameErrorRegex.FindStringSubmatch(line)
		if len(match) > 0 {
			errorCode, _ = strconv.Atoi(match[1])
			if currentObject != nil {
				data = append(data, currentObject)
			}
			continue
		}

		// Check for attribute
		match = objectAttrRegex.FindStringSubmatch(line)
		if len(match) > 0 && currentObject != nil {
			currentObject[match[1]] = match[2]
			continue
		}
	}

	return Response{
		Error: errorCode,
		Data:  data,
	}
}

// PrettifyResponse converts raw response data to proper types
func (p *Protocol) PrettifyResponse(resp Response) Response {
	intAttrs := map[string]bool{
		"index":      true,
		"sendResult": true,
	}
	boolAttrs := map[string]bool{
		"unread": true,
	}
	dateAttrs := map[string]bool{
		"receivedTime": true,
		"sendTime":     true,
	}

	for i, obj := range resp.Data {
		for key, val := range obj {
			if intAttrs[key] {
				if v, err := strconv.Atoi(val.(string)); err == nil {
					obj[key] = v
				}
			} else if boolAttrs[key] {
				if v, err := strconv.Atoi(val.(string)); err == nil {
					obj[key] = v > 0
				}
			} else if dateAttrs[key] {
				if t, err := time.Parse("2006-01-02 15:04:05", val.(string)); err == nil {
					obj[key] = t
				}
			} else if key == "content" {
				valStr := val.(string)
				valStr = strings.ReplaceAll(valStr, "\u0012", "\n")
				obj[key] = valStr
			}
		}
		resp.Data[i] = obj
	}

	return resp
}
