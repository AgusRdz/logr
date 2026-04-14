package formats

import (
	"strconv"
	"strings"
	"time"
)

// Entry is a parsed log line.
type Entry struct {
	Timestamp time.Time
	Level     string         // normalized: DEBUG INFO WARN ERROR FATAL
	Message   string
	Fields    map[string]any
	Raw       []byte
	ParseErr  bool
}

// Format probes and parses log lines.
type Format interface {
	Probe(line []byte) bool
	Parse(line []byte) Entry
}

// Shared field name aliases used by Generic and as fallbacks in other formats.
var LevelAliases = []string{"level", "lvl", "severity", "log_level"}
var MsgAliases = []string{"msg", "message", "text", "body"}
var TsAliases = []string{"ts", "time", "timestamp", "at", "created_at"}

// NormalizeLevel maps raw level strings to canonical uppercase forms.
func NormalizeLevel(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "TRACE", "TRC", "VERBOSE":
		return "DEBUG"
	case "DEBUG", "DBG", "D":
		return "DEBUG"
	case "INFO", "INFORMATION", "INF", "I":
		return "INFO"
	case "WARN", "WARNING", "WRN", "W":
		return "WARN"
	case "ERROR", "ERR", "E":
		return "ERROR"
	case "FATAL", "CRITICAL", "CRIT", "PANIC", "F":
		return "FATAL"
	case "":
		return "INFO"
	default:
		return strings.ToUpper(strings.TrimSpace(raw))
	}
}

// ParseTimestamp converts a value (float64 or string) to a time.Time.
func ParseTimestamp(v any) time.Time {
	switch t := v.(type) {
	case float64:
		if t > 1e12 {
			ms := int64(t)
			return time.Unix(ms/1000, (ms%1000)*int64(time.Millisecond)).UTC()
		}
		return time.Unix(int64(t), 0).UTC()
	case string:
		return ParseTimeString(t)
	}
	return time.Time{}
}

// ParseTimeString tries multiple layouts to parse a timestamp string.
func ParseTimeString(s string) time.Time {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	// Fall back to numeric string → ParseTimestamp
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return ParseTimestamp(f)
	}
	return time.Time{}
}
