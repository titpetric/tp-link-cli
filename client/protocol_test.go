package client

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeDataFrame(t *testing.T) {
	proto := NewProtocol()

	reqs := []Request{
		{
			Method:     ActGet,
			Controller: "LTE_SMS_RECVMSGENTRY",
			Attrs:      []string{"index", "from", "content"},
		},
	}

	frame := proto.MakeDataFrame(reqs)
	assert.NotEmpty(t, frame)
	assert.Contains(t, frame, "1")       // ActGet
	assert.Contains(t, frame, "LTE_SMS") // controller
}

func TestToKVWithString(t *testing.T) {
	proto := NewProtocol()

	result := proto.toKV("test string")
	assert.Equal(t, "test string", result)
}

func TestToKVWithArray(t *testing.T) {
	proto := NewProtocol()

	attrs := []string{"index", "from", "content"}
	result := proto.toKV(attrs)

	assert.Contains(t, result, "index")
	assert.Contains(t, result, "from")
	assert.Contains(t, result, "content")
	assert.Contains(t, result, "\r\n")
}

func TestToKVWithMap(t *testing.T) {
	proto := NewProtocol()

	attrs := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": nil,
	}

	result := proto.toKV(attrs)
	assert.Contains(t, result, "key1=value1")
	assert.Contains(t, result, "key2=42")
	assert.Contains(t, result, "key3\r\n")
}

func TestFromDataFrame(t *testing.T) {
	proto := NewProtocol()

	frame := `1
[0,0,0,0,0,0]0,2
index=1
from=+1234567890
[error]0`

	result := proto.FromDataFrame(frame)
	assert.Equal(t, 0, result.Error)
	assert.Equal(t, 1, len(result.Data))
	assert.Equal(t, "1", result.Data[0]["index"])
	assert.Equal(t, "+1234567890", result.Data[0]["from"])
}

func TestFromDataFrameWithError(t *testing.T) {
	proto := NewProtocol()

	frame := `1
[error]1`

	result := proto.FromDataFrame(frame)
	assert.Equal(t, 1, result.Error)
}

func TestPrettifyResponse(t *testing.T) {
	proto := NewProtocol()

	resp := Response{
		Error: 0,
		Data: []map[string]interface{}{
			{
				"index":   "42",
				"unread":  "1",
				"content": "test\u0012message",
			},
		},
	}

	prettified := proto.PrettifyResponse(resp)
	assert.Equal(t, 0, prettified.Error)
	assert.Equal(t, 42, prettified.Data[0]["index"])
	assert.Equal(t, true, prettified.Data[0]["unread"])
	assert.True(t, strings.Contains(prettified.Data[0]["content"].(string), "\n"))
}

func TestMakeDataFrameWithStack(t *testing.T) {
	proto := NewProtocol()

	reqs := []Request{
		{
			Method:     ActDel,
			Controller: "LTE_SMS_RECVMSGENTRY",
			Stack:      "5,0,0,0,0,0",
		},
	}

	frame := proto.MakeDataFrame(reqs)
	assert.Contains(t, frame, "5,0,0,0,0,0")
}

func TestToKVWithNewlineReplacement(t *testing.T) {
	proto := NewProtocol()

	attrs := map[string]interface{}{
		"content": "line1\nline2\rline3",
	}

	result := proto.toKV(attrs)
	assert.Contains(t, result, "line1\u0012line2\u0012line3")
}
