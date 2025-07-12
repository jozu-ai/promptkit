package daemon

import (
	"fmt"
	"log"
	"net/http"

	"github.com/promptkit/promptkit/internal/appdir"
	"github.com/promptkit/promptkit/internal/recorder"
)

// Run starts the promptkit daemon and blocks until the HTTP server exits.
func Run(addr, backend string) error {
	logPath, err := appdir.SessionLogPath()
	if err != nil {
		return fmt.Errorf("session path: %w", err)
	}

	rec, err := recorder.New(logPath)
	if err != nil {
		return fmt.Errorf("recorder: %w", err)
	}
	defer rec.Close()

	handler, err := newHandler(backend, rec)
	if err != nil {
		return fmt.Errorf("handler: %w", err)
	}

	log.Printf("promptkit listening on %s", addr)
	return http.ListenAndServe(addr, handler)
}
