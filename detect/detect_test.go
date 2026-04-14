package detect

import (
	"testing"

	"github.com/AgusRdz/logr/formats"
)

func TestDetect(t *testing.T) {
	pinoLines := [][]byte{
		[]byte(`{"v":1,"pid":42,"level":30,"msg":"hello","time":1700000000000}`),
		[]byte(`{"v":1,"pid":42,"level":40,"msg":"warn","time":1700000000000}`),
	}
	got := Detect(pinoLines)
	if _, ok := got.(formats.Pino); !ok {
		t.Errorf("pino lines: got %T, want formats.Pino", got)
	}

	winstonLines := [][]byte{
		[]byte(`{"level":"info","message":"hello","timestamp":"2024-01-15T10:30:00Z"}`),
		[]byte(`{"level":"error","message":"fail","timestamp":"2024-01-15T10:31:00Z"}`),
	}
	got2 := Detect(winstonLines)
	if _, ok := got2.(formats.Winston); !ok {
		t.Errorf("winston lines: got %T, want formats.Winston", got2)
	}

	lambdaLines := [][]byte{
		[]byte(`{"timestamp":"2024-01-15T10:30:00Z","message":"invoked","requestId":"abc"}`),
	}
	got3 := Detect(lambdaLines)
	if _, ok := got3.(formats.Lambda); !ok {
		t.Errorf("lambda lines: got %T, want formats.Lambda", got3)
	}

	// empty/non-JSON lines → Passthrough
	emptyLines := [][]byte{
		[]byte(``),
		[]byte(`plain text`),
	}
	got4 := Detect(emptyLines)
	if _, ok := got4.(formats.Passthrough); !ok {
		t.Errorf("empty lines: got %T, want formats.Passthrough", got4)
	}
}

func TestByName(t *testing.T) {
	cases := []struct {
		name string
		want interface{}
	}{
		{"pino", formats.Pino{}},
		{"winston", formats.Winston{}},
		{"lambda", formats.Lambda{}},
		{"cloudwatch", formats.CloudWatch{}},
		{"generic", formats.Generic{}},
		{"unknown", formats.Generic{}},
		{"", formats.Generic{}},
	}
	for _, c := range cases {
		got := ByName(c.name)
		if got != c.want {
			t.Errorf("ByName(%q) = %T, want %T", c.name, got, c.want)
		}
	}
}
