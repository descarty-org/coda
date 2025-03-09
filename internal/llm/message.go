package llm

import (
	"encoding/json"
	"time"
)

// MessageList contains a list of messages.
type MessageList struct {
	Messages []*Message `json:"messages"`
}

// Message represents a single message in a conversation.
type Message struct {
	ID           string        `json:"id,omitempty"`
	Hidden       bool          `json:"hidden,omitempty"`
	Role         Role          `json:"role"`
	Content      string        `json:"content"`
	Name         string        `json:"name,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
	FinishReason string        `json:"finish_reason,omitempty"`
	Completed    bool          `json:"completed,omitempty"`
	Error        *string       `json:"error,omitempty"`
	Length       int           `json:"length,omitempty"`
	Timestamp    time.Time     `json:"timestamp,omitempty"`
	Metadata     any           `json:"metadata,omitempty"`
}

// Role represents the role of a message sender.
type Role string

// Standard message roles
const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleFunction  Role = "function"
)

// NewMessage creates a new message with the given role and content.
func NewMessage(role Role, content string) Message {
	return Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}
}

// NewSystemMessage creates a new system message.
func NewSystemMessage(content string) Message {
	return NewMessage(RoleSystem, content)
}

// NewUserMessage creates a new user message.
func NewUserMessage(content string) Message {
	return NewMessage(RoleUser, content)
}

// NewAssistantMessage creates a new assistant message.
func NewAssistantMessage(content string) Message {
	return NewMessage(RoleAssistant, content)
}

// NewFunctionMessage creates a new function message.
func NewFunctionMessage(name string, content string) Message {
	msg := NewMessage(RoleFunction, content)
	msg.Name = name
	return msg
}

// IsError returns true if the message is an error.
func (m *Message) IsError() bool {
	return m.Error != nil
}

// IsFunctionCall returns true if the message is a function call.
func (m *Message) IsFunctionCall() bool {
	return m.FunctionCall != nil
}

// IsFunction returns true if the message is a function.
func (m *Message) IsFunction() bool {
	return m.Role == RoleFunction
}

// FunctionName returns the function name.
func (m *Message) FunctionName() string {
	if m.FunctionCall != nil && m.FunctionCall.Name != "" {
		return m.FunctionCall.Name
	}
	return m.Name
}

// IsCompleted returns true if the message is completed.
func (m *Message) IsCompleted() bool {
	return m.Completed || m.FinishReason != ""
}

// ToMap converts the message to a map for easier serialization.
func (m *Message) ToMap() map[string]any {
	result := map[string]any{
		"role":    m.Role,
		"content": m.Content,
	}

	if m.Name != "" {
		result["name"] = m.Name
	}

	if m.FunctionCall != nil {
		result["function_call"] = m.FunctionCall
	}

	return result
}

// MarshalJSON implements the json.Marshaler interface.
func (m *Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp,omitempty"`
	}{
		Alias:     (*Alias)(m),
		Timestamp: m.Timestamp.Format(time.RFC3339),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	aux := &struct {
		*Alias
		Timestamp string `json:"timestamp,omitempty"`
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Timestamp != "" {
		timestamp, err := time.Parse(time.RFC3339, aux.Timestamp)
		if err != nil {
			return err
		}
		m.Timestamp = timestamp
	}

	return nil
}

// FunctionCall represents a call to a function.
type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// ParseArguments parses the arguments string into the provided struct.
func (fc *FunctionCall) ParseArguments(v any) error {
	return json.Unmarshal([]byte(fc.Arguments), v)
}

// SetArguments sets the arguments from a struct.
func (fc *FunctionCall) SetArguments(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	fc.Arguments = string(data)
	return nil
}
