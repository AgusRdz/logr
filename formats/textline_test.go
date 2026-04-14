package formats

import "testing"

func TestTextlineProbe(t *testing.T) {
	tl := Textline{}

	cases := []struct {
		line string
		want bool
	}{
		{"2025-04-12 14:23:05 ERROR payment failed", true},
		{"2025-04-12T14:23:05.123Z INFO  server started", true},
		{"2025-04-12 14:23:05 WARN  slow query detected", true},
		{"ERROR: disk full", true},
		{"[WARN] connection timeout", false}, // starts with '[' - excluded
		{`{"level":"info","message":"json"}`, false},
		{`level=info msg="logfmt"`, false},
		{"just some plain text without level", false},
		{"the ERROR count increased to 5", false}, // no timestamp, level not at start
	}

	for _, tc := range cases {
		got := tl.Probe([]byte(tc.line))
		if got != tc.want {
			t.Errorf("Probe(%q) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestTextlineParse(t *testing.T) {
	tl := Textline{}

	e := tl.Parse([]byte("2025-04-12 14:23:05 ERROR payment failed orderId=99"))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "ERROR" {
		t.Errorf("Level = %q, want ERROR", e.Level)
	}
	if e.Message != "payment failed orderId=99" {
		t.Errorf("Message = %q, want 'payment failed orderId=99'", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestTextlineParseNoTimestamp(t *testing.T) {
	tl := Textline{}

	e := tl.Parse([]byte("ERROR: disk full"))
	if e.Level != "ERROR" {
		t.Errorf("Level = %q, want ERROR", e.Level)
	}
	if e.Message != "disk full" {
		t.Errorf("Message = %q, want 'disk full'", e.Message)
	}
}
