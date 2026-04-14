package render

import (
	"bytes"
	"testing"

	"github.com/AgusRdz/logr/formats"
)

func TestRawWrite(t *testing.T) {
	var buf bytes.Buffer
	r := NewRaw(&buf)
	raw := []byte(`{"level":"info","msg":"test"}`)
	e := formats.Entry{
		Raw:   raw,
		Level: "INFO",
	}
	r.Write(e)

	out := buf.String()
	expected := string(raw) + "\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}
