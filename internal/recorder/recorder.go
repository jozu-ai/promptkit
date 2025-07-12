package recorder

import (
	"encoding/json"
	"fmt"
	"os"
)

// Recorder writes sessions to a JSON Lines file.
type Recorder struct {
	file *os.File
}

// New creates a new Recorder writing to the given file path.
func New(path string) (*Recorder, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &Recorder{file: f}, nil
}

// Close closes the underlying file.
func (r *Recorder) Close() error {
	if r.file == nil {
		return nil
	}
	return r.file.Close()
}

// Record writes the given session object as JSON to the file.
func (r *Recorder) Record(v interface{}) error {
	if r.file == nil {
		return fmt.Errorf("recorder closed")
	}
	enc := json.NewEncoder(r.file)
	if err := enc.Encode(v); err != nil {
		return err
	}
	return nil
}
