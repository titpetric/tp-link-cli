package model

import "time"

// SMSMessage represents a single SMS message.
type SMSMessage struct {
	Index    int       `json:"index"`
	From     string    `json:"from,omitempty"`
	To       string    `json:"to,omitempty"`
	Content  string    `json:"content"`
	SentTime time.Time `json:"sendTime,omitempty"`
	RecvTime time.Time `json:"receivedTime,omitempty"`
	Unread   bool      `json:"unread,omitempty"`
}

// ListResponse is the response from List operation.
type ListResponse struct {
	Error int          `json:"error"`
	Data  []SMSMessage `json:"data"`
}

// ReadResponse is the response from Read operation.
type ReadResponse struct {
	Error int          `json:"error"`
	Data  []SMSMessage `json:"data"`
}

// DeleteResponse is the response from Delete operation.
type DeleteResponse struct {
	Error int          `json:"error"`
	Data  []SMSMessage `json:"data"`
}

// SendResponse is the response from Send operation.
type SendResponse struct {
	Error int          `json:"error"`
	Data  []SMSMessage `json:"data"`
}
