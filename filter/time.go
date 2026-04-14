package filter

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseSince parses a relative or absolute time string into a time.Time.
func ParseSince(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("cannot parse time %q", s)
	}

	// Relative: number + unit (s/m/h/d/w)
	if len(s) >= 2 {
		unit := s[len(s)-1]
		numStr := s[:len(s)-1]
		if n, err := strconv.ParseFloat(numStr, 64); err == nil {
			var d time.Duration
			switch unit {
			case 's':
				d = time.Duration(n * float64(time.Second))
			case 'm':
				d = time.Duration(n * float64(time.Minute))
			case 'h':
				d = time.Duration(n * float64(time.Hour))
			case 'd':
				d = time.Duration(n * float64(24*time.Hour))
			case 'w':
				d = time.Duration(n * float64(7*24*time.Hour))
			}
			if d > 0 {
				return time.Now().Add(-d), nil
			}
		}
	}

	// Time-only: "15:04:05" or "15:04" (length <= 8 and contains ":")
	if len(s) <= 8 && strings.Contains(s, ":") {
		now := time.Now()
		for _, layout := range []string{"15:04:05", "15:04"} {
			if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
				return time.Date(now.Year(), now.Month(), now.Day(),
					t.Hour(), t.Minute(), t.Second(), 0, time.Local), nil
			}
		}
	}

	// Absolute formats
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse time %q", s)
}

// ParseUntil is an alias for ParseSince - same parsing rules.
func ParseUntil(s string) (time.Time, error) { return ParseSince(s) }
