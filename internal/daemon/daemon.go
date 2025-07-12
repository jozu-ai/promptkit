package daemon

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/promptkit/promptkit/internal/proxy"
	"github.com/promptkit/promptkit/internal/recorder"
	"github.com/promptkit/promptkit/pkg/session"
)

// Run starts the promptkit daemon and blocks until the HTTP server exits.
func Run(addr, backend, logFile string) error {
	rec, err := recorder.New(logFile)
	if err != nil {
		return fmt.Errorf("recorder: %w", err)
	}
	defer rec.Close()

	rp, err := proxy.New(backend)
	if err != nil {
		return fmt.Errorf("proxy: %w", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rp.ServeHTTP(w, r)
		sess := session.Session{
			ID:       time.Now().Format("20060102150405"),
			Origin:   "manual",
			Request:  map[string]any{"method": r.Method, "path": r.URL.Path},
			Response: map[string]any{"status": w.Header().Get("Status")},
			Metadata: session.Metadata{
				Timestamp: time.Now(),
				LatencyMS: time.Since(start).Milliseconds(),
			},
		}
		if err := rec.Record(&sess); err != nil {
			log.Printf("record: %v", err)
		}
	})

	log.Printf("promptkit listening on %s", addr)
	return http.ListenAndServe(addr, handler)
}
