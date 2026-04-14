package formats

import "testing"

func TestWinstonProbe(t *testing.T) {
	w := Winston{}

	// level+message+timestamp → true
	if !w.Probe([]byte(`{"level":"info","message":"hello","timestamp":"2024-01-15T10:30:00Z"}`)) {
		t.Error("expected true for level+message+timestamp")
	}

	// missing timestamp → false
	if w.Probe([]byte(`{"level":"info","message":"hello"}`)) {
		t.Error("expected false for missing timestamp")
	}

	// missing message → false
	if w.Probe([]byte(`{"level":"info","timestamp":"2024-01-15T10:30:00Z"}`)) {
		t.Error("expected false for missing message")
	}

	// has "v" → false (Pino match guard)
	if w.Probe([]byte(`{"level":"info","message":"hello","timestamp":"2024-01-15T10:30:00Z","v":1}`)) {
		t.Error("expected false when 'v' present")
	}

	// invalid JSON → false
	if w.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestWinstonParse(t *testing.T) {
	w := Winston{}

	e := w.Parse([]byte(`{"level":"error","message":"something failed","timestamp":"2024-01-15T10:30:00Z","service":"api"}`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "ERROR" {
		t.Errorf("Level = %q, want ERROR", e.Level)
	}
	if e.Message != "something failed" {
		t.Errorf("Message = %q, want 'something failed'", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if e.Fields["service"] != "api" {
		t.Errorf("Fields[service] = %v, want api", e.Fields["service"])
	}
	for _, k := range []string{"level", "message", "timestamp"} {
		if _, ok := e.Fields[k]; ok {
			t.Errorf("field %q should be removed from Fields", k)
		}
	}
}
