package session

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// ComputeHash returns a deterministic hash of the session excluding the
// SessionHash field itself.
func ComputeHash(s Session) (string, error) {
	cpy := s
	cpy.Metadata.SessionHash = ""
	b, err := json.Marshal(cpy)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
