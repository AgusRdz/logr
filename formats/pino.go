package formats

import "encoding/json"

// Pino handles Node.js pino structured logs identified by "v" and "pid" fields.
type Pino struct{}

func (Pino) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	_, hasV := m["v"]
	_, hasPid := m["pid"]
	return hasV && hasPid
}

func (Pino) Parse(line []byte) Entry {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	// Level: integer encoding
	var level string
	if v, ok := m["level"]; ok {
		switch lv := v.(type) {
		case float64:
			switch {
			case lv < 20:
				level = "DEBUG"
			case lv < 30:
				level = "DEBUG"
			case lv < 40:
				level = "INFO"
			case lv < 50:
				level = "WARN"
			case lv < 60:
				level = "ERROR"
			default:
				level = "FATAL"
			}
		case string:
			level = NormalizeLevel(lv)
		}
		delete(m, "level")
	}

	// Message
	var msg string
	if v, ok := m["msg"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
		delete(m, "msg")
	}

	// Timestamp: Unix milliseconds
	var ts any
	if v, ok := m["time"]; ok {
		ts = v
		delete(m, "time")
	}

	// Remove noise fields
	delete(m, "v")
	delete(m, "pid")
	delete(m, "hostname")

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     level,
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}
