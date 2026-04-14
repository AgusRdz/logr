package stats

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/AgusRdz/logr/formats"
)

// Stats accumulates log entry counts.
type Stats struct {
	Total    int
	ByLevel  map[string]int
	ByMinute map[string]int // key: "15:04" in local time
	First    time.Time
	Last     time.Time
	errMsgs  map[string]int
}

func New() *Stats {
	return &Stats{
		ByLevel:  make(map[string]int),
		ByMinute: make(map[string]int),
		errMsgs:  make(map[string]int),
	}
}

func (s *Stats) Add(e formats.Entry) {
	s.Total++
	s.ByLevel[e.Level]++
	if !e.Timestamp.IsZero() {
		local := e.Timestamp.Local()
		if s.First.IsZero() || local.Before(s.First) {
			s.First = local
		}
		if s.Last.IsZero() || local.After(s.Last) {
			s.Last = local
		}
		s.ByMinute[local.Format("15:04")]++
	}
	if e.Level == "ERROR" || e.Level == "FATAL" {
		s.errMsgs[e.Message]++
	}
}

func (s *Stats) Print(w io.Writer, color bool) {
	fmt.Fprintln(w, "log stats")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  total:  %s entries\n", formatInt(s.Total))

	levelOrder := []string{"FATAL", "ERROR", "WARN", "INFO", "DEBUG"}

	// Find max count for bar scaling
	maxCount := 0
	for _, l := range levelOrder {
		if s.ByLevel[l] > maxCount {
			maxCount = s.ByLevel[l]
		}
	}

	for _, l := range levelOrder {
		count := s.ByLevel[l]
		if count == 0 {
			continue
		}
		pct := float64(count) / float64(s.Total) * 100
		bar := progressBar(count, maxCount, 20)
		label := levelLabelColored(l, color)
		fmt.Fprintf(w, "  %s: %s  (%.4g%%)   %s\n",
			label, formatInt(count), pct, bar)
	}

	// By-minute section
	if len(s.ByMinute) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  by minute (last 10m shown):")

		keys := make([]string, 0, len(s.ByMinute))
		for k := range s.ByMinute {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Show last 15 buckets
		if len(keys) > 15 {
			keys = keys[len(keys)-15:]
		}

		maxMin := 0
		for _, k := range keys {
			if s.ByMinute[k] > maxMin {
				maxMin = s.ByMinute[k]
			}
		}

		for _, k := range keys {
			bar := progressBar(s.ByMinute[k], maxMin, 16)
			fmt.Fprintf(w, "  %s %s %d\n", k, bar, s.ByMinute[k])
		}
	}

	// Top errors section
	if len(s.errMsgs) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  top errors:")

		type pair struct {
			msg   string
			count int
		}
		pairs := make([]pair, 0, len(s.errMsgs))
		for msg, cnt := range s.errMsgs {
			pairs = append(pairs, pair{msg, cnt})
		}
		sort.Slice(pairs, func(i, j int) bool {
			if pairs[i].count != pairs[j].count {
				return pairs[i].count > pairs[j].count
			}
			return pairs[i].msg < pairs[j].msg
		})

		limit := 5
		if len(pairs) < limit {
			limit = len(pairs)
		}
		for i := 0; i < limit; i++ {
			fmt.Fprintf(w, "  %d. %q   (%d times)\n", i+1, pairs[i].msg, pairs[i].count)
		}
	}
}

func progressBar(value, max, width int) string {
	if max == 0 {
		return strings.Repeat("░", width)
	}
	filled := int(float64(value)/float64(max)*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func formatInt(n int) string {
	s := fmt.Sprintf("%d", n)
	// Insert commas
	result := make([]byte, 0, len(s)+len(s)/3)
	offset := len(s) % 3
	if offset == 0 {
		offset = 3
	}
	result = append(result, s[:offset]...)
	for i := offset; i < len(s); i += 3 {
		result = append(result, ',')
		result = append(result, s[i:i+3]...)
	}
	return string(result)
}

func levelLabelColored(level string, color bool) string {
	if !color {
		return fmt.Sprintf("%-6s", level)
	}
	const (
		dim      = "\033[2m"
		yellow   = "\033[33m"
		red      = "\033[31m"
		redBold  = "\033[1;31m"
		reset    = "\033[0m"
	)
	label := fmt.Sprintf("%-6s", level)
	switch level {
	case "DEBUG":
		return dim + label + reset
	case "WARN":
		return yellow + label + reset
	case "ERROR":
		return red + label + reset
	case "FATAL":
		return redBold + label + reset
	default:
		return label
	}
}
