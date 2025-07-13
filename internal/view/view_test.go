package view

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/promptkit/promptkit/pkg/session"
)

func writeSession(t *testing.T, f *os.File, s session.Session) {
	t.Helper()
	enc := json.NewEncoder(f)
	if err := enc.Encode(s); err != nil {
		t.Fatal(err)
	}
}

func TestFindSession(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writeSession(t, f, session.Session{ID: "1", Metadata: session.Metadata{Timestamp: time.Now()}})
	writeSession(t, f, session.Session{ID: "2", Metadata: session.Metadata{Timestamp: time.Now()}})
	f.WriteString("{bad}\n")
	f.Close()

	s, err := FindSession(dir, "2")
	if err != nil {
		t.Fatal(err)
	}
	if s == nil || s.ID != "2" {
		t.Fatalf("unexpected session: %+v", s)
	}

	s, err = FindSession(dir, "missing")
	if err != nil {
		t.Fatal(err)
	}
	if s != nil {
		t.Fatalf("expected nil, got %+v", s)
	}
}
