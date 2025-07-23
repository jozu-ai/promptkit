package session

import "time"

type Origin string

const (
	OriginManual    Origin = "manual"
	OriginFramework Origin = "framework"
	OriginModelKit  Origin = "modelkit"
	OriginProxy     Origin = "proxy" // For backward compatibility with existing sessions
)

// Session represents one prompt-response transaction recorded by promptkit.
type Session struct {
	ID           string         `json:"id"`
	Origin       Origin         `json:"origin"`
	SourcePrompt string         `json:"source_prompt"`
	Request      OpenAIRequest  `json:"request"`
	Response     OpenAIResponse `json:"response"`
	Stream       bool           `json:"stream"`
	Metadata     Metadata       `json:"metadata"`
}

// Metadata holds auxiliary metadata about the session.
type Metadata struct {
	Timestamp   time.Time `json:"timestamp"`
	LatencyMS   int64     `json:"latency_ms"`
	Tags        []string  `json:"tags,omitempty"`
	Published   *string   `json:"published,omitempty"` // OCI ref if published
	SessionHash string    `json:"session_hash"`
}

// OpenAIRequest captures a prompt sent to the OpenAI-compatible API.
type OpenAIRequest struct {
	Model       string      `json:"model,omitempty"`
	Messages    []Message   `json:"messages,omitempty"` // chat models
	Prompt      interface{} `json:"prompt,omitempty"`   // non-chat models
	Temperature float64     `json:"temperature,omitempty"`
	TopP        float64     `json:"top_p,omitempty"`
	MaxTokens   int         `json:"max_tokens,omitempty"`
	Stop        interface{} `json:"stop,omitempty"`
	Stream      bool        `json:"stream,omitempty"`
	// Legacy fields for backward compatibility with existing proxy format
	Method  string      `json:"method,omitempty"`
	Path    string      `json:"path,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

// OpenAIResponse captures the response from the OpenAI-compatible API.
type OpenAIResponse struct {
	ID      string     `json:"id"`
	Object  string     `json:"object"`
	Created int64      `json:"created"` // Unix timestamp
	Model   string     `json:"model"`
	Choices []Choice   `json:"choices"`
	Usage   UsageStats `json:"usage,omitempty"`
	// Legacy fields for backward compatibility with existing proxy format
	Status int         `json:"status,omitempty"`
	Body   interface{} `json:"body,omitempty"`
}

// Message is a single role-content pair from a chat completion request or response.
type Message struct {
	Role    string `json:"role"`           // "system", "user", "assistant"
	Content string `json:"content"`        // The message text
	Name    string `json:"name,omitempty"` // Optional name for tool/function
}

// Choice represents one completion candidate.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"` // "stop", "length", etc.
}

// UsageStats contains token usage information.
type UsageStats struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
