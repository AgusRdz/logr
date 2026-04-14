package formats

import "testing"

func TestPassthroughProbe(t *testing.T) {
	p := Passthrough{}
	// always true
	if !p.Probe([]byte(`anything`)) {
		t.Error("expected true for any input")
	}
	if !p.Probe([]byte(`{"json":"line"}`)) {
		t.Error("expected true for JSON")
	}
	if !p.Probe([]byte(``)) {
		t.Error("expected true for empty")
	}
}

func TestPassthroughParse(t *testing.T) {
	p := Passthrough{}
	line := []byte(`some plain log line`)
	e := p.Parse(line)

	if !e.ParseErr {
		t.Error("expected ParseErr=true")
	}
	if string(e.Raw) != string(line) {
		t.Errorf("Raw = %q, want %q", e.Raw, line)
	}
	if e.Message != string(line) {
		t.Errorf("Message = %q, want %q", e.Message, string(line))
	}
}
