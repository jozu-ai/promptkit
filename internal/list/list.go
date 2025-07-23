package list

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/promptkit/promptkit/pkg/session"
)

// FilterFunc evaluates whether a session map matches.
type FilterFunc func(map[string]any) bool

var filterRe = regexp.MustCompile(`^\s*([\w\.]+)\s*(=|!=|~|<|>)\s*(\S+)\s*$`)

// ParseFilter parses a filter expression into a FilterFunc.
func ParseFilter(expr string) (FilterFunc, error) {
	if strings.TrimSpace(expr) == "" {
		return func(map[string]any) bool { return true }, nil
	}
	m := filterRe.FindStringSubmatch(expr)
	if m == nil {
		return nil, fmt.Errorf("invalid filter expression")
	}
	path := strings.Split(m[1], ".")
	op := m[2]
	valStr := m[3]

	return func(data map[string]any) bool {
		v, _ := getPathValue(data, path)
		return compare(v, op, valStr)
	}, nil
}

// LoadSessions reads all sessions from the given directory, sorted by timestamp descending.
func LoadSessions(dir string) ([]session.Session, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		return nil, err
	}
	var sessions []session.Session
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
			sessions = append(sessions, s)
		}
		if err := sc.Err(); err != nil {
			log.Printf("reading %s: %v", f, err)
		}
		file.Close()
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Metadata.Timestamp.After(sessions[j].Metadata.Timestamp)
	})
	return sessions, nil
}

// ToMap converts a Session into a generic map for filtering.
func ToMap(s session.Session) map[string]any {
	b, _ := json.Marshal(s)
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}

func getPathValue(data any, path []string) (any, bool) {
	if len(path) == 0 {
		return data, true
	}
	m, ok := data.(map[string]any)
	if !ok {
		return nil, false
	}
	v, ok := m[path[0]]
	if !ok {
		return nil, false
	}
	return getPathValue(v, path[1:])
}

func compare(v any, op, target string) bool {
	switch op {
	case "=":
		if target == "null" {
			return v == nil
		}
		return fmt.Sprint(v) == target
	case "!=":
		if target == "null" {
			return v != nil
		}
		return fmt.Sprint(v) != target
	case "~":
		if arr, ok := v.([]any); ok {
			for _, item := range arr {
				if fmt.Sprint(item) == target {
					return true
				}
			}
			return false
		}
		if v == nil {
			return false
		}
		return strings.Contains(fmt.Sprint(v), target)
	case ">", "<":
		if tStr, ok := v.(string); ok {
			if tVal, err := time.Parse(time.RFC3339, tStr); err == nil {
				if cmpVal, err2 := time.Parse(time.RFC3339, target); err2 == nil {
					if op == ">" {
						return tVal.After(cmpVal)
					}
					return tVal.Before(cmpVal)
				}
			}
		}
		num, ok1 := toFloat64(v)
		cmp, err2 := strconv.ParseFloat(target, 64)
		if ok1 && err2 == nil {
			if op == ">" {
				return num > cmp
			}
			return num < cmp
		}
		return false
	default:
		return false
	}
}

func toFloat64(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case string:
		if f, err := strconv.ParseFloat(t, 64); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// Summary represents a simplified session for output.
type Summary struct {
	ID        string   `json:"id"`
	Model     string   `json:"model"`
	Origin    string   `json:"origin"`
	Tokens    int      `json:"tokens"`
	LatencyMS int64    `json:"latency_ms"`
	Tags      []string `json:"tags"`
	Published string   `json:"published"`
}

// Summarize converts a session to a Summary.
func Summarize(s session.Session) Summary {
	m := ToMap(s)
	model, _ := getPathValue(m, []string{"request", "model"})
	if model == nil {
		model, _ = getPathValue(m, []string{"request", "payload", "model"})
	}
	tokensVal, _ := getPathValue(m, []string{"response", "usage", "total_tokens"})
	tokens, _ := toFloat64(tokensVal)

	pub := ""
	if s.Metadata.Published != nil {
		pub = *s.Metadata.Published
	}
	return Summary{
		ID:        s.ID,
		Model:     fmt.Sprint(model),
		Origin:    string(s.Origin),
		Tokens:    int(tokens),
		LatencyMS: s.Metadata.LatencyMS,
		Tags:      s.Metadata.Tags,
		Published: pub,
	}
}

// PrintTable prints summaries in a simple table format.
func PrintTable(summaries []Summary) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tModel\tOrigin\tTokens\tLatency\tTags\tPublished")
	fmt.Fprintln(w, "--\t-----\t------\t------\t-------\t----\t---------")
	for _, s := range summaries {
		tags := strings.Join(s.Tags, ", ")
		latency := fmt.Sprintf("%dms", s.LatencyMS)
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n", s.ID, s.Model, s.Origin, s.Tokens, latency, tags, s.Published)
	}
	w.Flush()
}
