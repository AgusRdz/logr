package formats

import "encoding/json"

// Serilog handles .NET Serilog Compact Log Event Format (CLEF).
// Identified by "@t" (timestamp) and "@mt" (message template) fields.
type Serilog struct{}

func (Serilog) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	_, hasT := m["@t"]
	_, hasMt := m["@mt"]
	return hasT && hasMt
}

func (Serilog) Parse(line []byte) Entry {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	// @l is optional; omitted level means Information
	var level string
	if v, ok := m["@l"]; ok {
		if s, ok := v.(string); ok {
			level = NormalizeLevel(s)
		}
		delete(m, "@l")
	} else {
		level = "INFO"
	}

	// @mt is the message template; use it as the message
	var msg string
	if v, ok := m["@mt"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
		delete(m, "@mt")
	}

	var ts any
	if v, ok := m["@t"]; ok {
		ts = v
		delete(m, "@t")
	}

	// @x is the exception — keep in fields
	// Remove other CLEF meta fields
	delete(m, "@i") // event id

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     level,
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}
