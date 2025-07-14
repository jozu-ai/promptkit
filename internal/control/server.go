package control

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
	"github.com/promptkit/promptkit/internal/appdir"
	"github.com/promptkit/promptkit/internal/list"
	"github.com/promptkit/promptkit/pkg/session"
	"github.com/promptkit/promptkit/pkg/version"
)

// Broker sends session events to subscribers.
type Broker struct {
	mu      sync.Mutex
	clients map[chan session.Session]struct{}
}

func newBroker() *Broker {
	return &Broker{clients: make(map[chan session.Session]struct{})}
}

func (b *Broker) Subscribe() chan session.Session {
	ch := make(chan session.Session, 10)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broker) Unsubscribe(ch chan session.Session) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	close(ch)
}

func (b *Broker) Broadcast(s session.Session) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		select {
		case ch <- s:
		default:
		}
	}
}

// Server holds state for the control-plane server.
type Server struct {
	addr   string
	dir    string
	broker *Broker
	http   *http.Server
}

// NewServer creates a new Server for the given address.
func NewServer(addr string) (*Server, error) {
	dir, err := appdir.SessionsDir()
	if err != nil {
		return nil, err
	}
	b := newBroker()
	srv := &Server{addr: addr, dir: dir, broker: b}

	r := chi.NewRouter()
	r.Get("/status", srv.handleStatus)
	r.Get("/sessions", srv.handleSessions)
	r.Get("/sessions/{id}", srv.handleSession)
	r.Get("/events", srv.handleEvents)
	srv.http = &http.Server{Addr: addr, Handler: r}
	return srv, nil
}

// Start begins watching sessions and serving HTTP.
func (s *Server) Start() error {
	if err := s.watch(); err != nil {
		return err
	}
	go func() {
		log.Printf("promptkit control server listening on %s", s.addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()
	return nil
}

func (s *Server) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := watcher.Add(s.dir); err != nil {
		return err
	}
	known := map[string]struct{}{}
	sessions, _ := list.LoadSessions(s.dir)
	for _, ss := range sessions {
		known[ss.ID] = struct{}{}
	}
	go func() {
		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				if ev.Op&(fsnotify.Create|fsnotify.Write) != 0 && strings.HasSuffix(ev.Name, ".jsonl") {
					sess, err := list.LoadSessions(s.dir)
					if err != nil {
						log.Printf("load sessions: %v", err)
						continue
					}
					for _, ss := range sess {
						if _, exists := known[ss.ID]; !exists {
							known[ss.ID] = struct{}{}
							s.broker.Broadcast(ss)
						}
					}
				}
			case err := <-watcher.Errors:
				if err != nil {
					log.Printf("watch error: %v", err)
				}
			}
		}
	}()
	return nil
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok", "version": version.Version}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := list.LoadSessions(s.dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pred, err := list.ParseFilter(r.URL.Query().Get("filter"))
	if err != nil {
		http.Error(w, "invalid filter", http.StatusBadRequest)
		return
	}
	var summaries []list.Summary
	for _, ss := range sessions {
		if pred(list.ToMap(ss)) {
			summaries = append(summaries, list.Summarize(ss))
		}
	}
	if limStr := r.URL.Query().Get("limit"); limStr != "" {
		if lim, err := strconv.Atoi(limStr); err == nil && lim > 0 && lim < len(summaries) {
			summaries = summaries[:lim]
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sessions, err := list.LoadSessions(s.dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, ss := range sessions {
		if ss.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ss)
			return
		}
	}
	http.NotFound(w, r)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	ch := s.broker.Subscribe()
	defer s.broker.Unsubscribe(ch)

	notify := r.Context().Done()
	for {
		select {
		case s := <-ch:
			b, _ := json.Marshal(s)
			fmt.Fprintf(w, "event: session\n")
			fmt.Fprintf(w, "data: %s\n\n", b)
			flusher.Flush()
		case <-notify:
			return
		}
	}
}
