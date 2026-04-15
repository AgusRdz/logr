package formats

import (
	"encoding/json"
	"strings"
)

// Zap handles Go uber-go/zap structured logs.
// Identified by a float "ts" field, a "caller" field (file:line), and a "msg" field.
// Must be probed before Generic since it has "level" + "msg".
type Zap struct{}

func (Zap) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	ts, hasTs := m["ts"]
	if !hasTs {
		return false
	}
	// ts must be a float (Unix seconds with fractional part) not a large int string
	if _, ok := ts.(float64); !ok {
		return false
	}
	caller, hasCaller := m["caller"]
	if !hasCaller {
		return false
	}
	// caller looks like "pkg/file.go:42"
	if s, ok := caller.(string); !ok || !strings.Contains(s, ":") {
		return false
	}
	_, hasMsg := m["msg"]
	return hasMsg
}

func (Zap) Parse(line []byte) Entry {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	var level string
	if v, ok := m["level"]; ok {
		if s, ok := v.(string); ok {
			level = NormalizeLevel(s)
		}
		delete(m, "level")
	}

	var msg string
	if v, ok := m["msg"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
		delete(m, "msg")
	}

	var ts any
	if v, ok := m["ts"]; ok {
		ts = v
		delete(m, "ts")
	}

	// Keep caller in fields — it's useful signal
	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     level,
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}
