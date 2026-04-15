package formats

import (
	"bytes"
	"regexp"
)

// Rails handles Ruby on Rails production.log plain-text format.
// Format: "X, [TIMESTAMP #PID]  LEVEL -- TAG: message"
// Example: "I, [2025-04-12T14:23:05.123456 #1234]  INFO -- : Started GET /users"
type Rails struct{}

// railsRe matches the Rails logger format and captures timestamp, level, and message.
var railsRe = regexp.MustCompile(`^[A-Z], \[([^\]#]+)#\d+\]\s+([A-Z]+) -- [^:]*: (.+)$`)

// railsLevelMap maps single-char Rails level codes to canonical names.
var railsLevelMap = map[byte]string{
	'D': "DEBUG",
	'I': "INFO",
	'W': "WARN",
	'E': "ERROR",
	'F': "FATAL",
	'U': "ERROR", // UNKNOWN → ERROR
}

func (Rails) Probe(line []byte) bool {
	if len(line) < 5 {
		return false
	}
	// Must start with a single letter, comma, space, bracket
	if line[1] != ',' || line[2] != ' ' || line[3] != '[' {
		return false
	}
	if _, ok := railsLevelMap[line[0]]; !ok {
		return false
	}
	return railsRe.Match(line)
}

func (Rails) Parse(line []byte) Entry {
	m := railsRe.FindSubmatch(line)
	if m == nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	// Level from the spelled-out word in the match (e.g. "INFO"), not the single char
	level := NormalizeLevel(string(bytes.TrimSpace(m[2])))
	ts := ParseTimeString(string(bytes.TrimSpace(m[1])))
	msg := string(m[3])

	return Entry{
		Timestamp: ts,
		Level:     level,
		Message:   msg,
		Fields:    map[string]any{},
		Raw:       line,
	}
}
