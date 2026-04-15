package formats

import "encoding/json"

// SpringBoot handles Spring Boot JSON logs produced by logstash-logback-encoder.
// Identified by "@timestamp" + "level" + "message" fields.
// Must be probed before Generic since it overlaps on "level" + "message".
type SpringBoot struct{}

func (SpringBoot) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	_, hasTs := m["@timestamp"]
	_, hasLevel := m["level"]
	_, hasMsg := m["message"]
	return hasTs && hasLevel && hasMsg
}

func (SpringBoot) Parse(line []byte) Entry {
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

	var ts any
	if v, ok := m["@timestamp"]; ok {
		ts = v
		delete(m, "@timestamp")
	}

	// Remove logstash-logback-encoder structural noise
	delete(m, "@version")
	delete(m, "level_value")

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     level,
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}
