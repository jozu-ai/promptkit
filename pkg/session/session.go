package session

import "time"

// Session represents one prompt-response transaction recorded by promptkit.
type Session struct {
	ID           string      `json:"id"`
	Origin       string      `json:"origin"`
	SourcePrompt string      `json:"source_prompt"`
	Request      interface{} `json:"request"`
	Response     interface{} `json:"response"`
	Metadata     Metadata    `json:"metadata"`
}

type Metadata struct {
	Timestamp   time.Time `json:"timestamp"`
	LatencyMS   int64     `json:"latency_ms"`
	Tags        []string  `json:"tags"`
	Published   *string   `json:"published"`
	SessionHash string    `json:"session_hash"`
}
