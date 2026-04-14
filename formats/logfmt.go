package formats

import "strings"

// Logfmt handles logfmt-encoded log lines: key=value key="quoted value" ...
// Common in Go applications using logrus (default output) or go-kit/log.
type Logfmt struct{}

func (Logfmt) Probe(line []byte) bool {
	if len(line) == 0 || line[0] == '{' || line[0] == '[' {
		return false
	}
	s := string(line)
	// Require at least one recognized field with an explicit = assignment.
	known := append(append(LevelAliases, MsgAliases...), TsAliases...)
	for _, k := range known {
		if strings.Contains(s, k+"=") {
			return true
		}
	}
	return false
}

func (Logfmt) Parse(line []byte) Entry {
	pairs := parseLogfmtPairs(string(line))
	if len(pairs) == 0 {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	var level, msg string
	var ts any

	for _, alias := range LevelAliases {
		if v, ok := pairs[alias]; ok {
			level = v
			delete(pairs, alias)
			break
		}
	}
	for _, alias := range MsgAliases {
		if v, ok := pairs[alias]; ok {
			msg = v
			delete(pairs, alias)
			break
		}
	}
	for _, alias := range TsAliases {
		if v, ok := pairs[alias]; ok {
			ts = v
			delete(pairs, alias)
			break
		}
	}

	fields := make(map[string]any, len(pairs))
	for k, v := range pairs {
		fields[k] = v
	}

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     NormalizeLevel(level),
		Message:   msg,
		Fields:    fields,
		Raw:       line,
	}
}

// parseLogfmtPairs parses a logfmt line into key=value pairs.
func parseLogfmtPairs(s string) map[string]string {
	result := make(map[string]string)
	s = strings.TrimSpace(s)
	for len(s) > 0 {
		s = strings.TrimLeft(s, " \t")
		if s == "" {
			break
		}
		// Find key boundary
		i := 0
		for i < len(s) && s[i] != '=' && s[i] != ' ' && s[i] != '\t' {
			i++
		}
		if i == 0 {
			break
		}
		key := s[:i]
		s = s[i:]
		if len(s) == 0 || s[0] != '=' {
			result[key] = "true"
			continue
		}
		s = s[1:] // skip '='
		var value string
		if len(s) > 0 && s[0] == '"' {
			end := 1
			for end < len(s) {
				if s[end] == '\\' {
					end += 2
					continue
				}
				if s[end] == '"' {
					end++
					break
				}
				end++
			}
			raw := s[1 : end-1]
			raw = strings.ReplaceAll(raw, `\"`, `"`)
			raw = strings.ReplaceAll(raw, `\\`, `\`)
			value = raw
			s = s[end:]
		} else {
			j := 0
			for j < len(s) && s[j] != ' ' && s[j] != '\t' {
				j++
			}
			value = s[:j]
			s = s[j:]
		}
		result[key] = value
	}
	return result
}
