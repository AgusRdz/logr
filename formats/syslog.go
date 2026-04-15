package formats

import (
	"bytes"
	"regexp"
	"strconv"
)

// Syslog handles RFC 5424 syslog messages.
// Format: <PRIVAL>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID STRUCTURED-DATA MSG
// Example: <34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - message
type Syslog struct{}

// rfc5424 matches <priority>version timestamp and captures the rest.
var rfc5424 = regexp.MustCompile(`^<(\d{1,3})>(\d+) (\S+) \S+ \S+ \S+ \S+ [^\s]+ (.*)$`)

// priToLevel maps syslog severity (prival & 0x7) to canonical levels.
func priToLevel(prival int) string {
	switch prival & 0x7 {
	case 0: // EMERG
		return "FATAL"
	case 1: // ALERT
		return "FATAL"
	case 2: // CRIT
		return "FATAL"
	case 3: // ERR
		return "ERROR"
	case 4: // WARNING
		return "WARN"
	case 5: // NOTICE
		return "INFO"
	case 6: // INFO
		return "INFO"
	case 7: // DEBUG
		return "DEBUG"
	default:
		return "INFO"
	}
}

func (Syslog) Probe(line []byte) bool {
	if !bytes.HasPrefix(line, []byte("<")) {
		return false
	}
	return rfc5424.Match(line)
}

func (Syslog) Parse(line []byte) Entry {
	m := rfc5424.FindSubmatch(line)
	if m == nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	prival, _ := strconv.Atoi(string(m[1]))
	level := priToLevel(prival)
	ts := ParseTimeString(string(m[3]))
	msg := string(m[4])

	return Entry{
		Timestamp: ts,
		Level:     level,
		Message:   msg,
		Fields:    map[string]any{},
		Raw:       line,
	}
}
