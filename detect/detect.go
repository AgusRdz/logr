package detect

import "github.com/AgusRdz/logr/formats"

// ordered is the probe sequence - first match wins per line.
// More-specific formats must appear before less-specific ones.
// Passthrough is always last.
var ordered = []formats.Format{
	formats.Bunyan{},    // Node.js bunyan — before Pino (both have v+pid)
	formats.Pino{},      // Node.js pino
	formats.Winston{},   // Node.js winston
	formats.Lambda{},    // AWS Lambda
	formats.CloudWatch{}, // AWS CloudWatch
	formats.Serilog{},    // .NET Serilog CLEF (@t + @mt)
	formats.Log4j{},      // Java Log4j2 (timeMillis + loggerName)
	formats.Monolog{},    // PHP Monolog (level_name + channel)
	formats.Python{},     // Python logging/structlog (levelname or event+level)
	formats.Zap{},        // Go uber-go/zap (ts float + caller) — before Generic
	formats.SpringBoot{}, // Spring Boot / logstash-logback-encoder (@timestamp + level + message)
	formats.NginxJSON{},  // Nginx JSON access log (remote_addr + request + status)
	formats.Generic{},    // any JSON with level+msg
	formats.Logfmt{},    // key=value (Go logrus text, etc.)
	formats.Syslog{},    // RFC 5424 syslog (<prival>)
	formats.Rails{},     // Ruby on Rails production.log text
	formats.Npm{},       // npm debug logs (<seq> <level> <message>)
	formats.CLF{},       // Apache/Nginx access logs
	formats.Textline{},  // plain text with level prefix
	formats.Passthrough{},
}

// Detect votes on sample lines and returns the most-matched format.
// Non-Passthrough formats compete; Passthrough wins only if nothing else matches.
func Detect(lines [][]byte) formats.Format {
	return DetectWithExtras(lines, nil)
}

// DetectWithExtras is like Detect but prepends extras to the probe list so that
// user-defined formats are tried before built-in ones.
func DetectWithExtras(lines [][]byte, extras []formats.Format) formats.Format {
	candidates := make([]formats.Format, 0, len(extras)+len(ordered)-1)
	candidates = append(candidates, extras...)
	candidates = append(candidates, ordered[:len(ordered)-1]...) // exclude Passthrough

	votes := make([]int, len(candidates))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		for i, f := range candidates {
			if f.Probe(line) {
				votes[i]++
				break
			}
		}
	}

	best := -1
	bestVotes := 0
	for i, v := range votes {
		if v > bestVotes {
			bestVotes = v
			best = i
		}
	}

	if best < 0 {
		return formats.Passthrough{}
	}
	return candidates[best]
}

// ByName returns a Format by hint name.
// Returns Generic{} for unknown names.
func ByName(name string) formats.Format {
	switch name {
	case "pino":
		return formats.Pino{}
	case "bunyan":
		return formats.Bunyan{}
	case "winston":
		return formats.Winston{}
	case "lambda":
		return formats.Lambda{}
	case "cloudwatch":
		return formats.CloudWatch{}
	case "serilog":
		return formats.Serilog{}
	case "log4j", "log4j2":
		return formats.Log4j{}
	case "monolog":
		return formats.Monolog{}
	case "python":
		return formats.Python{}
	case "zap":
		return formats.Zap{}
	case "generic":
		return formats.Generic{}
	case "logfmt":
		return formats.Logfmt{}
	case "syslog":
		return formats.Syslog{}
	case "springboot", "spring-boot":
		return formats.SpringBoot{}
	case "nginxjson", "nginx-json":
		return formats.NginxJSON{}
	case "rails":
		return formats.Rails{}
	case "npm":
		return formats.Npm{}
	case "clf":
		return formats.CLF{}
	case "textline":
		return formats.Textline{}
	default:
		return formats.Generic{}
	}
}
