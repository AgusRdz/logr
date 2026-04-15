package formats

import "encoding/json"

// Python handles two common Python JSON log formats:
//   - stdlib logging with JsonFormatter: "levelname" + "msg" fields
//   - structlog JSON: "event" + "level" fields
//
// Must be probed before Generic.
type Python struct{}

func (Python) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	// stdlib logging JSON: has "levelname"
	if _, ok := m["levelname"]; ok {
		return true
	}
	// structlog JSON: has "event" as the message field alongside a level key
	if _, ok := m["event"]; ok {
		for _, alias := range LevelAliases {
			if _, ok := m[alias]; ok {
				return true
			}
		}
	}
	return false
}

func (Python) Parse(line []byte) Entry {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	var level string

	// stdlib logging uses "levelname"
	if v, ok := m["levelname"]; ok {
		if s, ok := v.(string); ok {
			level = NormalizeLevel(s)
		}
		delete(m, "levelname")
	} else {
		// structlog uses standard level aliases
		for _, alias := range LevelAliases {
			if v, ok := m[alias]; ok {
				if s, ok := v.(string); ok {
					level = NormalizeLevel(s)
				}
				delete(m, alias)
				break
			}
		}
	}

	var msg string
	// structlog uses "event" as the message key
	if v, ok := m["event"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
		delete(m, "event")
	} else if v, ok := m["msg"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
		delete(m, "msg")
	} else if v, ok := m["message"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
		delete(m, "message")
	}

	// Timestamp: stdlib uses "created" (Unix float) or "asctime" (string)
	var ts any
	for _, alias := range []string{"timestamp", "created", "asctime", "time"} {
		if v, ok := m[alias]; ok {
			ts = v
			delete(m, alias)
			break
		}
	}

	// Remove stdlib noise
	delete(m, "name")
	delete(m, "pathname")
	delete(m, "filename")
	delete(m, "module")
	delete(m, "funcName")
	delete(m, "lineno")
	delete(m, "process")
	delete(m, "processName")
	delete(m, "thread")
	delete(m, "threadName")

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     level,
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}
