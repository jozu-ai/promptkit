package list

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/promptkit/promptkit/pkg/session"
)

func TestParseFilter(t *testing.T) {
	pred, err := ParseFilter("request.model=gpt-4")
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]any{"request": map[string]any{"model": "gpt-4"}}
	if !pred(m) {
		t.Fatalf("expected match")
	}
}

func TestFilterOperators(t *testing.T) {
	pred, _ := ParseFilter("metadata.latency_ms<1000")
	m := map[string]any{"metadata": map[string]any{"latency_ms": 500.0}}
	if !pred(m) {
		t.Fatalf("latency predicate failed")
	}

	pred, _ = ParseFilter("metadata.timestamp>2025-07-01T00:00:00Z")
	m = map[string]any{"metadata": map[string]any{"timestamp": "2025-07-06T12:42:00Z"}}
	if !pred(m) {
		t.Fatalf("timestamp predicate failed")
	}

	pred, _ = ParseFilter("metadata.tags~qa")
	m = map[string]any{"metadata": map[string]any{"tags": []any{"qa", "other"}}}
	if !pred(m) {
		t.Fatalf("tags predicate failed")
	}

	pred, _ = ParseFilter("metadata.published!=null")
	m = map[string]any{"metadata": map[string]any{"published": "x"}}
	if !pred(m) {
		t.Fatalf("published predicate failed")
	}
}

func TestSummarize(t *testing.T) {
	pub := "oci://reg/app:1"
	s := session.Session{
		ID:     "1",
		Origin: session.OriginModelKit,
		Request: session.OpenAIRequest{
			Model: "gpt",
		},
		Response: session.OpenAIResponse{
			Usage: session.UsageStats{TotalTokens: 5},
		},
		Metadata: session.Metadata{Timestamp: time.Now(), LatencyMS: 10, Tags: []string{"a"}, Published: &pub},
	}
	sum := Summarize(s)
	if sum.Model != "gpt" || sum.Tokens != 5 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}

func TestToMap(t *testing.T) {
	s := session.Session{
		ID: "1",
		Request: session.OpenAIRequest{
			Model: "gpt",
		},
	}
	m := ToMap(s)
	b, _ := json.Marshal(m)
	if string(b) == "" {
		t.Fatalf("unexpected empty map")
	}
}
