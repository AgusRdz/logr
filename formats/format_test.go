package formats

import (
	"testing"
	"time"
)

func TestNormalizeLevel(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"TRACE", "DEBUG"}, {"TRC", "DEBUG"}, {"VERBOSE", "DEBUG"},
		{"DEBUG", "DEBUG"}, {"DBG", "DEBUG"}, {"D", "DEBUG"},
		{"debug", "DEBUG"}, {"dbg", "DEBUG"},
		{"INFO", "INFO"}, {"INFORMATION", "INFO"}, {"INF", "INFO"}, {"I", "INFO"},
		{"info", "INFO"},
		{"WARN", "WARN"}, {"WARNING", "WARN"}, {"WRN", "WARN"}, {"W", "WARN"},
		{"warn", "WARN"},
		{"ERROR", "ERROR"}, {"ERR", "ERROR"}, {"E", "ERROR"},
		{"error", "ERROR"},
		{"FATAL", "FATAL"}, {"CRITICAL", "FATAL"}, {"CRIT", "FATAL"}, {"PANIC", "FATAL"}, {"F", "FATAL"},
		{"fatal", "FATAL"},
		{"", "INFO"},
		{"CUSTOM", "CUSTOM"},
		{"  info  ", "INFO"},
	}
	for _, c := range cases {
		got := NormalizeLevel(c.input)
		if got != c.want {
			t.Errorf("NormalizeLevel(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestParseTimestamp(t *testing.T) {
	// float64 seconds
	ts := ParseTimestamp(float64(1700000000))
	if ts.IsZero() {
		t.Error("expected non-zero for unix seconds")
	}
	if ts.Unix() != 1700000000 {
		t.Errorf("unix seconds: got %d, want 1700000000", ts.Unix())
	}

	// float64 milliseconds (> 1e12)
	tsMs := ParseTimestamp(float64(1700000000000))
	if tsMs.IsZero() {
		t.Error("expected non-zero for unix milliseconds")
	}
	if tsMs.Unix() != 1700000000 {
		t.Errorf("unix ms: got %d, want 1700000000", tsMs.Unix())
	}

	// RFC3339 string
	tsStr := ParseTimestamp("2024-01-15T10:30:00Z")
	if tsStr.IsZero() {
		t.Error("expected non-zero for RFC3339 string")
	}
	if tsStr.Year() != 2024 {
		t.Errorf("RFC3339: year = %d, want 2024", tsStr.Year())
	}

	// Unix string
	tsUnixStr := ParseTimestamp("1700000000")
	if tsUnixStr.IsZero() {
		t.Error("expected non-zero for unix string")
	}
	if tsUnixStr.Unix() != 1700000000 {
		t.Errorf("unix string: got %d, want 1700000000", tsUnixStr.Unix())
	}
}

func TestParseTimeString(t *testing.T) {
	// RFC3339
	ts := ParseTimeString("2024-01-15T10:30:00Z")
	if ts.IsZero() || ts.Year() != 2024 {
		t.Errorf("RFC3339 parse failed: %v", ts)
	}

	// Custom format
	ts2 := ParseTimeString("2024-01-15T10:30:00")
	if ts2.IsZero() {
		t.Errorf("custom format parse failed: %v", ts2)
	}

	// Numeric string
	ts3 := ParseTimeString("1700000000")
	if ts3.IsZero() || ts3.Unix() != 1700000000 {
		t.Errorf("numeric string parse failed: %v", ts3)
	}

	// Unrecognized → zero
	ts4 := ParseTimeString("not-a-time")
	if !ts4.IsZero() {
		t.Errorf("unrecognized should return zero, got %v", ts4)
	}

	// Slash format
	ts5 := ParseTimeString("2024/01/15 10:30:00")
	if ts5.IsZero() {
		t.Errorf("slash format parse failed: %v", ts5)
	}

	// Ensure zero value is actually zero
	var zero time.Time
	if !zero.IsZero() == false {
		// zero.IsZero() should be true
	}
}
