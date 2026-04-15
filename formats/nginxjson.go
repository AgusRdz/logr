package formats

import "encoding/json"

// NginxJSON handles Nginx access logs configured with a JSON log_format.
// Identified by "remote_addr" + "request" + "status" fields.
// Must be probed before Generic.
type NginxJSON struct{}

func (NginxJSON) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	_, hasAddr := m["remote_addr"]
	_, hasReq := m["request"]
	_, hasStatus := m["status"]
	return hasAddr && hasReq && hasStatus
}

func (NginxJSON) Parse(line []byte) Entry {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	// Derive level from HTTP status code
	var level string
	if v, ok := m["status"]; ok {
		level = httpStatusLevel(v)
		delete(m, "status")
	}

	// Message is the request line
	var msg string
	if v, ok := m["request"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
		delete(m, "request")
	}

	// Timestamp: try common Nginx time field names
	var ts any
	for _, key := range []string{"time_iso8601", "time_local", "msec"} {
		if v, ok := m[key]; ok {
			ts = v
			delete(m, key)
			break
		}
	}

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     level,
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}

// httpStatusLevel maps an HTTP status code value to a canonical log level.
func httpStatusLevel(v any) string {
	switch s := v.(type) {
	case float64:
		n := int(s)
		switch {
		case n >= 500:
			return "ERROR"
		case n >= 400:
			return "WARN"
		default:
			return "INFO"
		}
	case string:
		if len(s) == 0 {
			return "INFO"
		}
		switch s[0] {
		case '5':
			return "ERROR"
		case '4':
			return "WARN"
		default:
			return "INFO"
		}
	}
	return "INFO"
}
