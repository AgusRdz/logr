package detect

import "github.com/AgusRdz/logr/formats"

// ordered is the probe sequence - first match wins per line.
// Passthrough is always last.
var ordered = []formats.Format{
	formats.Pino{},
	formats.Winston{},
	formats.Lambda{},
	formats.CloudWatch{},
	formats.Generic{},
	formats.Logfmt{},
	formats.CLF{},
	formats.Textline{},
	formats.Passthrough{},
}

// Detect votes on sample lines and returns the most-matched format.
// Non-Passthrough formats compete; Passthrough wins only if nothing else matches.
func Detect(lines [][]byte) formats.Format {
	votes := make([]int, len(ordered)-1) // exclude Passthrough

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		for i, f := range ordered[:len(ordered)-1] {
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
	return ordered[best]
}

// ByName returns a Format by hint name.
// Returns Generic{} for unknown names.
func ByName(name string) formats.Format {
	switch name {
	case "pino":
		return formats.Pino{}
	case "winston":
		return formats.Winston{}
	case "lambda":
		return formats.Lambda{}
	case "cloudwatch":
		return formats.CloudWatch{}
	case "generic":
		return formats.Generic{}
	case "logfmt":
		return formats.Logfmt{}
	case "clf":
		return formats.CLF{}
	case "textline":
		return formats.Textline{}
	default:
		return formats.Generic{}
	}
}
