package formats

import "encoding/json"

// Log4j handles Java Log4j2 JSON layout output.
// Identified by "timeMillis" and "loggerName" fields.
// Must be probed before Generic since it has "level" + "message".
type Log4j struct{}

func (Log4j) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	_, hasMillis := m["timeMillis"]
	_, hasLogger := m["loggerName"]
	return hasMillis && hasLogger
}

func (Log4j) Parse(line []byte) Entry {
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
	if v, ok := m["message"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
		delete(m, "message")
	}

	// timeMillis is Unix milliseconds as a float64
	var ts any
	if v, ok := m["timeMillis"]; ok {
		ts = v
		delete(m, "timeMillis")
	}

	// Remove noisy structural fields
	delete(m, "loggerFqcn")
	delete(m, "endOfBatch")
	delete(m, "contextMap")

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     level,
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}
