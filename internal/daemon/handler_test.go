package daemon

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/promptkit/promptkit/internal/recorder"
	"github.com/promptkit/promptkit/pkg/session"
)

func readSessions(t *testing.T, path string) []session.Session {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var sess []session.Session
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var s session.Session
		if err := json.Unmarshal(sc.Bytes(), &s); err != nil {
			t.Fatal(err)
		}
		sess = append(sess, s)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return sess
}

func TestRecordCompletions(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"1","object":"completion"}`)
	}))
	defer backend.Close()

	tmp, err := os.CreateTemp(t.TempDir(), "log")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	rec, err := recorder.New(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Close()

	h, err := newHandler(backend.URL, rec)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	_, err = http.Post(srv.URL+"/v1/completions", "application/json", strings.NewReader(`{"model":"gpt","prompt":"hi"}`))
	if err != nil {
		t.Fatal(err)
	}

	sess := readSessions(t, tmp.Name())
	if len(sess) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sess))
	}
	if sess[0].Stream {
		t.Fatalf("expected non-stream session")
	}
	reqMap := sess[0].Request.(map[string]any)
	if reqMap["path"] != "/v1/completions" {
		t.Fatalf("unexpected path: %v", reqMap["path"])
	}
	if sess[0].Metadata.SessionHash == "" {
		t.Fatalf("hash not set")
	}
}

func TestIgnoreOtherEndpoints(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{}`)
	}))
	defer backend.Close()

	tmp, _ := os.CreateTemp(t.TempDir(), "log")
	tmp.Close()

	rec, _ := recorder.New(tmp.Name())
	defer rec.Close()

	h, _ := newHandler(backend.URL, rec)
	srv := httptest.NewServer(h)
	defer srv.Close()

	http.Get(srv.URL + "/v1/models")
	http.Post(srv.URL+"/v1/embeddings", "application/json", strings.NewReader(`{}`))

	sess := readSessions(t, tmp.Name())
	if len(sess) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sess))
	}
}

func TestStreamFlag(t *testing.T) {
	// Backend that sends simple SSE stream
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hi\"}}]}\n\n")
		io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer backend.Close()

	tmp, _ := os.CreateTemp(t.TempDir(), "log")
	tmp.Close()
	rec, _ := recorder.New(tmp.Name())
	defer rec.Close()

	h, _ := newHandler(backend.URL, rec)
	srv := httptest.NewServer(h)
	defer srv.Close()

	http.Post(srv.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{"stream":true}`))

	sess := readSessions(t, tmp.Name())
	if len(sess) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sess))
	}
	if !sess[0].Stream {
		t.Fatalf("expected stream flag true")
	}
}
