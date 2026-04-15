package formats

import "encoding/json"

// Bunyan handles Node.js bunyan structured logs.
// Identified by "v":0, "name", and "pid" fields with an ISO8601 "time" string.
// Must be probed before Pino since both have "v" and "pid".
type Bunyan struct{}

func (Bunyan) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	_, hasName := m["name"]
	_, hasPid := m["pid"]
	if !hasName || !hasPid {
		return false
	}
	// Bunyan "time" is always an ISO8601 string; Pino uses a numeric millisecond epoch.
	t, hasTime := m["time"]
	if !hasTime {
		return false
	}
	_, timeIsString := t.(string)
	return timeIsString
}

func (Bunyan) Parse(line []byte) Entry {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	// Level: numeric encoding identical to Pino (10=TRACE, 20=DEBUG, 30=INFO, 40=WARN, 50=ERROR, 60=FATAL)
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

	var msg string
	if v, ok := m["msg"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
		delete(m, "msg")
	}

	var ts any
	if v, ok := m["time"]; ok {
		ts = v
		delete(m, "time")
	}

	// Remove noise fields
	delete(m, "v")
	delete(m, "pid")
	delete(m, "hostname")
	delete(m, "name")

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     level,
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}
