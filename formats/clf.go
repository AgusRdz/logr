package formats

import (
	"regexp"
	"strconv"
	"time"
)

// clfRe matches Common Log Format and Combined Log Format lines.
// 127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /path HTTP/1.1" 200 1234
// Combined adds: "referrer" "user-agent"
var clfRe = regexp.MustCompile(
	`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) (\S+)[^"]*" (\d{3}) (\S+)` +
		`(?:\s+"([^"]*)" "([^"]*)")?`,
)

// CLF handles Apache/Nginx Common Log Format and Combined Log Format.
type CLF struct{}

func (CLF) Probe(line []byte) bool {
	return clfRe.Match(line)
}

func (CLF) Parse(line []byte) Entry {
	m := clfRe.FindSubmatch(line)
	if m == nil {
		return Entry{ParseErr: true, Raw: line, Message: string(line)}
	}

	host := string(m[1])
	timeStr := string(m[2])
	method := string(m[3])
	path := string(m[4])
	statusStr := string(m[5])
	sizeStr := string(m[6])
	referrer := string(m[7])
	agent := string(m[8])

	ts, _ := time.Parse("02/Jan/2006:15:04:05 -0700", timeStr)
	status, _ := strconv.Atoi(statusStr)

	var level string
	switch {
	case status >= 500:
		level = "ERROR"
	case status >= 400:
		level = "WARN"
	default:
		level = "INFO"
	}

	fields := map[string]any{
		"host":   host,
		"method": method,
		"path":   path,
		"status": status,
	}
	if sizeStr != "-" {
		if size, err := strconv.Atoi(sizeStr); err == nil {
			fields["size"] = size
		}
	}
	if referrer != "" && referrer != "-" {
		fields["referrer"] = referrer
	}
	if agent != "" && agent != "-" {
		fields["agent"] = agent
	}

	return Entry{
		Timestamp: ts,
		Level:     level,
		Message:   method + " " + path,
		Fields:    fields,
		Raw:       line,
	}
}
