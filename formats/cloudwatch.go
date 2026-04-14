package formats

import "encoding/json"

// CloudWatch handles AWS CloudWatch logs with "logEvents" array or "logGroup"+"logStream".
type CloudWatch struct{}

func (CloudWatch) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	if _, ok := m["logEvents"]; ok {
		return true
	}
	_, hasGroup := m["logGroup"]
	_, hasStream := m["logStream"]
	return hasGroup && hasStream
}

func (CloudWatch) Parse(line []byte) Entry {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	if eventsRaw, ok := m["logEvents"]; ok {
		events, ok := eventsRaw.([]any)
		if !ok || len(events) == 0 {
			return Entry{ParseErr: true, Raw: line, Message: string(line)}
		}
		first, ok := events[0].(map[string]any)
		if !ok {
			return Entry{ParseErr: true, Raw: line, Message: string(line)}
		}

		var ts any
		if t, ok := first["timestamp"]; ok {
			ts = t
		}

		var msg string
		var level string
		fields := map[string]any{}

		if msgRaw, ok := first["message"]; ok {
			if msgStr, ok := msgRaw.(string); ok {
				// Try to parse embedded JSON
				var inner map[string]any
				if err := json.Unmarshal([]byte(msgStr), &inner); err == nil {
					innerEntry := (Generic{}).Parse([]byte(msgStr))
					if !innerEntry.ParseErr {
						if !innerEntry.Timestamp.IsZero() {
							ts = innerEntry.Timestamp
						}
						if logGroup, ok := m["logGroup"]; ok {
							innerEntry.Fields["logGroup"] = logGroup
						}
						innerEntry.Raw = line
						return innerEntry
					}
				}
				msg = msgStr
			}
		}

		if logGroup, ok := m["logGroup"]; ok {
			fields["logGroup"] = logGroup
		}

		if level == "" {
			level = "INFO"
		}

		return Entry{
			Timestamp: ParseTimestamp(ts),
			Level:     level,
			Message:   msg,
			Fields:    fields,
			Raw:       line,
		}
	}

	// Fallback: parse like Generic
	entry := (Generic{}).Parse(line)
	entry.Raw = line
	return entry
}
