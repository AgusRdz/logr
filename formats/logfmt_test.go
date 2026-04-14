package formats

import "testing"

func TestLogfmtProbe(t *testing.T) {
	lf := Logfmt{}

	cases := []struct {
		line string
		want bool
	}{
		{`level=info msg="server started" time=2025-04-12T14:23:05Z port=3000`, true},
		{`level=error msg="payment failed" requestId=abc123`, true},
		{`ts=2025-04-12T14:23:05Z level=warn msg="slow query"`, true},
		{`{"level":"info","message":"json"}`, false},        // JSON
		{`not key=value pairs here`, false},                 // no recognized field
		{`foo=bar baz=qux`, false},                          // no recognized field
		{`127.0.0.1 - - [12/Apr/2025:14:23:05 +0000] "GET / HTTP/1.1" 200 1234`, false}, // CLF
	}

	for _, c := range cases {
		got := lf.Probe([]byte(c.line))
		if got != c.want {
			t.Errorf("Probe(%q) = %v, want %v", c.line, got, c.want)
		}
	}
}

func TestLogfmtParse(t *testing.T) {
	lf := Logfmt{}

	e := lf.Parse([]byte(`level=error msg="payment failed" time=2025-04-12T14:23:05Z requestId=abc123 latency=42ms`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "ERROR" {
		t.Errorf("Level = %q, want ERROR", e.Level)
	}
	if e.Message != "payment failed" {
		t.Errorf("Message = %q, want 'payment failed'", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if e.Fields["requestId"] != "abc123" {
		t.Errorf("Fields[requestId] = %v, want abc123", e.Fields["requestId"])
	}
	if e.Fields["latency"] != "42ms" {
		t.Errorf("Fields[latency] = %v, want 42ms", e.Fields["latency"])
	}
	for _, k := range []string{"level", "msg", "time"} {
		if _, ok := e.Fields[k]; ok {
			t.Errorf("field %q should be removed from Fields", k)
		}
	}
}

func TestLogfmtParseQuotedValues(t *testing.T) {
	lf := Logfmt{}

	e := lf.Parse([]byte(`level=info msg="a \"quoted\" value" time=2025-04-12T14:23:05Z`))
	if e.Message != `a "quoted" value` {
		t.Errorf("Message = %q, want 'a \"quoted\" value'", e.Message)
	}
}
