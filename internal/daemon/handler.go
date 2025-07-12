package daemon

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/promptkit/promptkit/internal/recorder"
	"github.com/promptkit/promptkit/pkg/session"
)

// newHandler returns an HTTP handler that proxies requests to the backend and
// records sessions for supported endpoints.
func newHandler(backend string, rec *recorder.Recorder) (http.Handler, error) {
	base, err := url.Parse(backend)
	if err != nil {
		return nil, err
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		r.Body.Close()

		// Forward the request to the backend.
		targetURL := base.ResolveReference(r.URL)
		req, err := http.NewRequest(r.Method, targetURL.String(), bytes.NewReader(bodyBytes))
		if err != nil {
			http.Error(w, "proxy error", http.StatusInternalServerError)
			return
		}
		req.Header = r.Header.Clone()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(respBody)

		// Determine if this request should be recorded.
		path := r.URL.Path
		if r.Method != http.MethodPost || (path != "/v1/completions" && path != "/v1/chat/completions") {
			return
		}

		var reqPayload map[string]any
		if err := json.Unmarshal(bodyBytes, &reqPayload); err != nil {
			return // malformed payload
		}
		var respPayload any
		if err := json.Unmarshal(respBody, &respPayload); err != nil {
			respPayload = string(respBody)
		}

		stream := false
		if v, ok := reqPayload["stream"].(bool); ok && v {
			stream = true
		}

		sess := session.Session{
			ID:           time.Now().Format("20060102150405"),
			Origin:       "proxy",
			SourcePrompt: "",
			Request: map[string]any{
				"method":  r.Method,
				"path":    path,
				"payload": reqPayload,
			},
			Response: map[string]any{
				"status": resp.StatusCode,
				"body":   respPayload,
			},
			Stream: stream,
			Metadata: session.Metadata{
				Timestamp: time.Now(),
				LatencyMS: time.Since(start).Milliseconds(),
			},
		}

		hash, err := session.ComputeHash(sess)
		if err == nil {
			sess.Metadata.SessionHash = hash
		}

		if err := rec.Record(&sess); err != nil {
			log.Printf("record: %v", err)
		}
	}), nil
}
