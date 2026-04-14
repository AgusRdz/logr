package formats

import "encoding/json"

// Winston handles Node.js winston logs identified by "level" + "message" + "timestamp" (no "v").
type Winston struct{}

func (Winston) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	if _, hasV := m["v"]; hasV {
		return false
	}
	_, hasLevel := m["level"]
	_, hasMessage := m["message"]
	_, hasTimestamp := m["timestamp"]
	return hasLevel && hasMessage && hasTimestamp
}

func (Winston) Parse(line []byte) Entry {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	var level, msg string
	var ts any

	if v, ok := m["level"]; ok {
		if s, ok := v.(string); ok {
			level = s
		}
		delete(m, "level")
	}
	if v, ok := m["message"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
		delete(m, "message")
	}
	if v, ok := m["timestamp"]; ok {
		ts = v
		delete(m, "timestamp")
	}

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     NormalizeLevel(level),
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}
