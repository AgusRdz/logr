package formats

import (
	"regexp"
	"strings"
)

// textTsRe matches a leading ISO-8601 or space-separated timestamp.
var textTsRe = regexp.MustCompile(
	`^\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:[.,]\d+)?(?:Z|[+-]\d{2}:\d{2})?`,
)

// textLevelRe matches a level keyword, optionally bracketed or followed by colon.
var textLevelRe = regexp.MustCompile(
	`(?i)^\[?(DEBUG|INFO|WARN(?:ING)?|ERROR|FATAL|CRITICAL|TRACE)\]?:?\s*`,
)

// Textline handles generic timestamped plain-text log lines.
// Examples:
//
//	2025-04-12 14:23:05 ERROR payment failed
//	2025-04-12T14:23:05.123Z INFO  server started
//	ERROR: disk full
//	[WARN] connection timeout
type Textline struct{}

func (Textline) Probe(line []byte) bool {
	if len(line) == 0 || line[0] == '{' || line[0] == '[' {
		return false
	}
	s := strings.TrimSpace(string(line))
	// Must start with a timestamp or a level keyword.
	if textTsRe.MatchString(s) {
		rest := strings.TrimSpace(s[textTsRe.FindStringIndex(s)[1]:])
		return textLevelRe.MatchString(rest)
	}
	return textLevelRe.MatchString(s)
}

func (Textline) Parse(line []byte) Entry {
	s := strings.TrimSpace(string(line))

	var ts any
	if loc := textTsRe.FindStringIndex(s); loc != nil && loc[0] == 0 {
		ts = s[loc[0]:loc[1]]
		s = strings.TrimSpace(s[loc[1]:])
	}

	var level string
	if m := textLevelRe.FindStringSubmatch(s); m != nil {
		level = m[1] // capture group: just the keyword, no brackets or colon
		s = strings.TrimSpace(s[len(m[0]):])
	}

	return Entry{
		Timestamp: ParseTimestamp(ts),
		Level:     NormalizeLevel(level),
		Message:   s,
		Fields:    map[string]any{},
		Raw:       line,
	}
}
