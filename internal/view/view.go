package view

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/promptkit/promptkit/pkg/session"
)

// FindSession searches all .jsonl files under dir for a session with the given ID.
func FindSession(dir, id string) (*session.Session, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			return nil, err
		}
		sc := bufio.NewScanner(file)
		for sc.Scan() {
			var s session.Session
			if err := json.Unmarshal(sc.Bytes(), &s); err != nil {
				log.Printf("skip corrupt session in %s: %v", f, err)
				continue
			}
			if s.ID == id {
				file.Close()
				return &s, nil
			}
		}
		if err := sc.Err(); err != nil {
			log.Printf("reading %s: %v", f, err)
		}
		file.Close()
	}
	return nil, nil
}
