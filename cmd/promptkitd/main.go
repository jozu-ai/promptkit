package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/promptkit/promptkit/internal/proxy"
	"github.com/promptkit/promptkit/internal/recorder"
	"github.com/promptkit/promptkit/pkg/session"
)

func main() {
	var (
		addr    = flag.String("addr", ":8080", "listen address")
		backend = flag.String("backend", "https://api.openai.com", "backend base URL")
		logFile = flag.String("log", "sessions.jsonl", "session log file")
	)
	flag.Parse()

	rec, err := recorder.New(*logFile)
	if err != nil {
		log.Fatalf("recorder: %v", err)
	}
	defer rec.Close()

	proxy, err := proxy.New(*backend)
	if err != nil {
		log.Fatalf("proxy: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		proxy.ServeHTTP(w, r)
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

	log.Printf("promptkitd listening on %s", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal(err)
	}
}
