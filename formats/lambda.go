package formats

import "encoding/json"

// Lambda handles AWS Lambda JSON logs with "timestamp" + "message", no "v".
type Lambda struct{}

func (Lambda) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	if _, hasV := m["v"]; hasV {
		return false
	}
	_, hasTimestamp := m["timestamp"]
	_, hasMessage := m["message"]
	return hasTimestamp && hasMessage
}

func (Lambda) Parse(line []byte) Entry {
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

	normalizedLevel := NormalizeLevel(level)
	if normalizedLevel == "" {
		normalizedLevel = "INFO"
	}

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     normalizedLevel,
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}
