package formats

import "encoding/json"

// Generic handles any JSON line that has at least one level-like and one message-like field.
type Generic struct{}

func (Generic) Probe(line []byte) bool {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	hasLevel := false
	for _, alias := range LevelAliases {
		if _, ok := m[alias]; ok {
			hasLevel = true
			break
		}
	}
	hasMsg := false
	for _, alias := range MsgAliases {
		if _, ok := m[alias]; ok {
			hasMsg = true
			break
		}
	}
	return hasLevel && hasMsg
}

func (Generic) Parse(line []byte) Entry {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	var level, msg string
	var ts any

	for _, alias := range LevelAliases {
		if v, ok := m[alias]; ok {
			if s, ok := v.(string); ok {
				level = s
			}
			delete(m, alias)
			break
		}
	}
	for _, alias := range MsgAliases {
		if v, ok := m[alias]; ok {
			if s, ok := v.(string); ok {
				msg = s
			}
			delete(m, alias)
			break
		}
	}
	for _, alias := range TsAliases {
		if v, ok := m[alias]; ok {
			ts = v
			delete(m, alias)
			break
		}
	}

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     NormalizeLevel(level),
		Message:   msg,
		Fields:    m,
		Raw:       line,
	}
}
