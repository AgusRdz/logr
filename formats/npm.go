package formats

import (
	"bytes"
	"regexp"
)

// Npm handles npm debug log files.
// Format: "<seq> <level> <component> <message>"
// Example: "12 silly package data  name: 'runjs',"
type Npm struct{}

// npmLevels is the set of known npm log level words used for probing.
var npmLevels = map[string]bool{
	"silly": true, "verbose": true, "info": true, "timing": true,
	"http": true, "notice": true, "warn": true, "error": true,
	"pause": true, "resume": true, "silent": true,
}

// npmRe matches "<seq> <level> <rest>" — seq is discarded, rest is the message.
var npmRe = regexp.MustCompile(`^\d+ (\w+) (.+)$`)

func (Npm) Probe(line []byte) bool {
	m := npmRe.FindSubmatch(line)
	if m == nil {
		return false
	}
	return npmLevels[string(bytes.ToLower(m[1]))]
}

func (Npm) Parse(line []byte) Entry {
	m := npmRe.FindSubmatch(line)
	if m == nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	level := NormalizeLevel(string(m[1]))
	msg := string(m[2])

	return Entry{
		Level:   level,
		Message: msg,
		Fields:  map[string]any{},
		Raw:     line,
	}
}
