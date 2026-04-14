package formats

import "testing"

func TestGenericProbe(t *testing.T) {
	g := Generic{}

	// valid JSON with level+msg
	if !g.Probe([]byte(`{"level":"info","msg":"hello"}`)) {
		t.Error("expected true for level+msg JSON")
	}

	// missing level
	if g.Probe([]byte(`{"msg":"hello"}`)) {
		t.Error("expected false for missing level")
	}

	// missing msg
	if g.Probe([]byte(`{"level":"info"}`)) {
		t.Error("expected false for missing msg")
	}

	// invalid JSON
	if g.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}

	// lvl alias
	if !g.Probe([]byte(`{"lvl":"info","msg":"hello"}`)) {
		t.Error("expected true for lvl alias")
	}

	// message alias
	if !g.Probe([]byte(`{"level":"info","message":"hello"}`)) {
		t.Error("expected true for message alias")
	}
}

func TestGenericParse(t *testing.T) {
	g := Generic{}

	line := []byte(`{"level":"warn","msg":"something bad","ts":1700000000,"request_id":"abc123"}`)
	e := g.Parse(line)

	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "WARN" {
		t.Errorf("Level = %q, want WARN", e.Level)
	}
	if e.Message != "something bad" {
		t.Errorf("Message = %q, want 'something bad'", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if e.Fields["request_id"] != "abc123" {
		t.Errorf("Fields[request_id] = %v, want abc123", e.Fields["request_id"])
	}
	if _, ok := e.Fields["level"]; ok {
		t.Error("level should be removed from Fields")
	}
	if _, ok := e.Fields["msg"]; ok {
		t.Error("msg should be removed from Fields")
	}

	// non-JSON
	e2 := g.Parse([]byte(`plain text`))
	if !e2.ParseErr {
		t.Error("expected ParseErr for non-JSON")
	}
}
