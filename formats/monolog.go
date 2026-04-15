package formats

import "encoding/json"

// Monolog handles PHP Monolog JSON output.
// Identified by "level_name", "channel", and "datetime" fields.
// Must be probed before Generic since it also has "message".
type Monolog struct{}

func (Monolog) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	_, hasLevelName := m["level_name"]
	_, hasChannel := m["channel"]
	_, hasDatetime := m["datetime"]
	return hasLevelName && hasChannel && hasDatetime
}

func (Monolog) Parse(line []byte) Entry {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	// Use level_name (string) over level (int code)
	var level string
	if v, ok := m["level_name"]; ok {
		if s, ok := v.(string); ok {
			level = NormalizeLevel(s)
		}
		delete(m, "level_name")
	}
	delete(m, "level") // numeric code, redundant

	var msg string
	if v, ok := m["message"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
		delete(m, "message")
	}

	var ts any
	if v, ok := m["datetime"]; ok {
		ts = v
		delete(m, "datetime")
	}

	// Flatten context and extra into fields, then remove wrappers
	if ctx, ok := m["context"].(map[string]any); ok {
		for k, v := range ctx {
			m["context."+k] = v
		}
	}
	delete(m, "context")

	if extra, ok := m["extra"].(map[string]any); ok {
		for k, v := range extra {
			m["extra."+k] = v
		}
	}
	delete(m, "extra")

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     level,
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}
