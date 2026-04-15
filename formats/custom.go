package formats

import (
	"encoding/json"
	"regexp"
)

// CustomFormatDef describes a user-defined log format.
// Two modes are supported — JSON field mapping and regex pattern — and they are
// mutually exclusive: if Pattern is set it takes precedence.
//
// JSON mode example (config.json):
//
//	"custom_formats": {
//	  "myapp": {
//	    "probe_field": "log_time",
//	    "ts_field":    "log_time",
//	    "level_field": "log_level",
//	    "msg_field":   "log_message",
//	    "level_map":   {"30": "INFO", "40": "WARN", "50": "ERROR"}
//	  }
//	}
//
// Regex mode example:
//
//	"rails": {
//	  "pattern": "^(?P<ts>\\S+) (?P<level>[A-Z]+) (?P<msg>.+)$"
//	}
//
// In regex mode, named groups "ts", "level", and "msg" are extracted.
// Any other named group is placed into Fields.
type CustomFormatDef struct {
	// JSON mode
	ProbeField string            `json:"probe_field"` // field whose presence identifies this format
	TSField    string            `json:"ts_field"`
	LevelField string            `json:"level_field"`
	MsgField   string            `json:"msg_field"`
	LevelMap   map[string]string `json:"level_map"` // raw value → canonical level

	// Regex mode (non-JSON lines)
	Pattern string `json:"pattern"` // named-group regex with optional ts/level/msg groups
}

// Custom is a Format built from a CustomFormatDef.
type Custom struct {
	Def CustomFormatDef
	re  *regexp.Regexp // compiled Pattern, nil when using JSON mode
}

// NewCustom compiles the Pattern (if any) and returns a ready Custom format.
// Returns an error only if Pattern is non-empty and fails to compile.
func NewCustom(def CustomFormatDef) (Custom, error) {
	c := Custom{Def: def}
	if def.Pattern != "" {
		re, err := regexp.Compile(def.Pattern)
		if err != nil {
			return Custom{}, err
		}
		c.re = re
	}
	return c, nil
}

func (c Custom) Probe(line []byte) bool {
	if c.re != nil {
		return c.re.Match(line)
	}
	if c.Def.ProbeField == "" {
		return false // no probe_field → only usable with explicit --format
	}
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return false
	}
	_, ok := m[c.Def.ProbeField]
	return ok
}

func (c Custom) Parse(line []byte) Entry {
	if c.re != nil {
		return c.parseRegex(line)
	}
	return c.parseJSON(line)
}

func (c Custom) parseJSON(line []byte) Entry {
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	var level, msg string
	var ts any

	if c.Def.LevelField != "" {
		if v, ok := m[c.Def.LevelField]; ok {
			switch lv := v.(type) {
			case string:
				level = lv
			case float64:
				// check level_map first (e.g. "30" → "INFO")
				key := formatFloat(lv)
				if mapped, ok := c.Def.LevelMap[key]; ok {
					level = mapped
				} else {
					level = numericLevel(lv)
				}
			}
			delete(m, c.Def.LevelField)
		}
	}
	if level != "" {
		// Apply string-level LevelMap if defined
		if mapped, ok := c.Def.LevelMap[level]; ok {
			level = mapped
		}
	}

	if c.Def.MsgField != "" {
		if v, ok := m[c.Def.MsgField]; ok {
			if s, ok := v.(string); ok {
				msg = s
			}
			delete(m, c.Def.MsgField)
		}
	}

	if c.Def.TSField != "" {
		if v, ok := m[c.Def.TSField]; ok {
			ts = v
			delete(m, c.Def.TSField)
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

func (c Custom) parseRegex(line []byte) Entry {
	m := c.re.FindSubmatch(line)
	if m == nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	fields := make(map[string]any)
	var tsStr, level, msg string

	for i, name := range c.re.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		val := string(m[i])
		switch name {
		case "ts":
			tsStr = val
		case "level":
			level = val
		case "msg":
			msg = val
		default:
			fields[name] = val
		}
	}

	if mapped, ok := c.Def.LevelMap[level]; ok {
		level = mapped
	}

	return Entry{
		Timestamp: ParseTimeString(tsStr),
		Level:     NormalizeLevel(level),
		Message:   msg,
		Fields:    fields,
		Raw:       line,
	}
}

// formatFloat converts a float64 to a string key for LevelMap lookup ("30", "200", etc.)
func formatFloat(f float64) string {
	i := int64(f)
	if float64(i) == f {
		s := make([]byte, 0, 8)
		if i < 0 {
			s = append(s, '-')
			i = -i
		}
		// itoa inline to avoid importing strconv in this path
		var buf [20]byte
		pos := len(buf)
		if i == 0 {
			s = append(s, '0')
		} else {
			for i > 0 {
				pos--
				buf[pos] = byte('0' + i%10)
				i /= 10
			}
			s = append(s, buf[pos:]...)
		}
		return string(s)
	}
	return ""
}

// numericLevel maps Pino/Bunyan-style integer levels to canonical strings.
func numericLevel(v float64) string {
	switch {
	case v < 30:
		return "DEBUG"
	case v < 40:
		return "INFO"
	case v < 50:
		return "WARN"
	case v < 60:
		return "ERROR"
	default:
		return "FATAL"
	}
}
