package filter

import (
	"strings"
	"testing"
	"time"
)

func TestParseSince(t *testing.T) {
	before := time.Now()

	cases := []struct {
		input    string
		minAgo   time.Duration
		maxAgo   time.Duration
	}{
		{"30s", 29 * time.Second, 31 * time.Second},
		{"10m", 9*time.Minute + 59*time.Second, 10*time.Minute + 1*time.Second},
		{"1h", 59*time.Minute + 59*time.Second, 1*time.Hour + 1*time.Second},
		{"7d", 7*24*time.Hour - time.Second, 7*24*time.Hour + time.Second},
		{"2w", 14*24*time.Hour - time.Second, 14*24*time.Hour + time.Second},
	}

	for _, c := range cases {
		ts, err := ParseSince(c.input)
		if err != nil {
			t.Errorf("ParseSince(%q) error: %v", c.input, err)
			continue
		}
		ago := before.Sub(ts)
		if ago < c.minAgo || ago > c.maxAgo {
			t.Errorf("ParseSince(%q): ago=%v, want [%v, %v]", c.input, ago, c.minAgo, c.maxAgo)
		}
	}
}

func TestParseAbsolute(t *testing.T) {
	// RFC3339
	ts, err := ParseSince("2024-01-15T10:30:00Z")
	if err != nil {
		t.Fatalf("RFC3339 error: %v", err)
	}
	if ts.Year() != 2024 || ts.Month() != 1 || ts.Day() != 15 {
		t.Errorf("RFC3339: got %v", ts)
	}

	// Date only
	ts2, err := ParseSince("2024-03-20")
	if err != nil {
		t.Fatalf("date-only error: %v", err)
	}
	if ts2.Year() != 2024 || ts2.Month() != 3 || ts2.Day() != 20 {
		t.Errorf("date-only: got %v", ts2)
	}

	// Time only
	ts3, err := ParseSince("14:30:00")
	if err != nil {
		t.Fatalf("time-only error: %v", err)
	}
	now := time.Now()
	if ts3.Year() != now.Year() || ts3.Month() != now.Month() || ts3.Day() != now.Day() {
		t.Errorf("time-only: expected today's date, got %v", ts3)
	}
	if ts3.Hour() != 14 || ts3.Minute() != 30 {
		t.Errorf("time-only: expected 14:30, got %02d:%02d", ts3.Hour(), ts3.Minute())
	}

	// Time only short
	ts4, err := ParseSince("09:15")
	if err != nil {
		t.Fatalf("short time-only error: %v", err)
	}
	if ts4.Hour() != 9 || ts4.Minute() != 15 {
		t.Errorf("short time-only: expected 09:15, got %02d:%02d", ts4.Hour(), ts4.Minute())
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := ParseSince("not-a-time")
	if err == nil {
		t.Error("expected error for invalid input")
	}
	if !strings.Contains(err.Error(), "cannot parse time") {
		t.Errorf("error message = %q, want 'cannot parse time'", err.Error())
	}

	_, err2 := ParseSince("")
	if err2 == nil {
		t.Error("expected error for empty string")
	}
}
