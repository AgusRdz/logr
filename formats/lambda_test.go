package formats

import "testing"

func TestLambdaProbe(t *testing.T) {
	l := Lambda{}

	// timestamp+message → true
	if !l.Probe([]byte(`{"timestamp":"2024-01-15T10:30:00Z","message":"hello"}`)) {
		t.Error("expected true for timestamp+message")
	}

	// has v → false
	if l.Probe([]byte(`{"timestamp":"2024-01-15T10:30:00Z","message":"hello","v":1}`)) {
		t.Error("expected false when 'v' present")
	}

	// missing message → false
	if l.Probe([]byte(`{"timestamp":"2024-01-15T10:30:00Z"}`)) {
		t.Error("expected false for missing message")
	}

	// missing timestamp → false
	if l.Probe([]byte(`{"message":"hello"}`)) {
		t.Error("expected false for missing timestamp")
	}

	// invalid JSON → false
	if l.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestLambdaParse(t *testing.T) {
	l := Lambda{}

	// with level
	e := l.Parse([]byte(`{"timestamp":"2024-01-15T10:30:00Z","message":"invoked","level":"warn","requestId":"xyz"}`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "WARN" {
		t.Errorf("Level = %q, want WARN", e.Level)
	}
	if e.Message != "invoked" {
		t.Errorf("Message = %q, want invoked", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if e.Fields["requestId"] != "xyz" {
		t.Errorf("Fields[requestId] = %v, want xyz", e.Fields["requestId"])
	}

	// missing level → default INFO
	e2 := l.Parse([]byte(`{"timestamp":"2024-01-15T10:30:00Z","message":"cold start"}`))
	if e2.Level != "INFO" {
		t.Errorf("missing level: got %q, want INFO", e2.Level)
	}

	// non-JSON
	e3 := l.Parse([]byte(`plain text`))
	if !e3.ParseErr {
		t.Error("expected ParseErr for non-JSON")
	}
}
