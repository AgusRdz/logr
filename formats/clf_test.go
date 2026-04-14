package formats

import "testing"

func TestCLFProbe(t *testing.T) {
	c := CLF{}

	cases := []struct {
		line string
		want bool
	}{
		{`127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`, true},
		{`192.168.1.1 - - [12/Apr/2025:14:23:05 +0000] "POST /api/login HTTP/1.1" 401 512`, true},
		{`10.0.0.1 - - [12/Apr/2025:14:23:05 +0000] "GET /health HTTP/1.1" 200 18 "https://example.com" "Mozilla/5.0"`, true},
		{`{"level":"info","message":"json"}`, false},
		{`level=info msg="logfmt"`, false},
		{`2025-04-12 14:23:05 INFO server started`, false},
	}

	for _, tc := range cases {
		got := c.Probe([]byte(tc.line))
		if got != tc.want {
			t.Errorf("Probe(%q) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestCLFParse(t *testing.T) {
	c := CLF{}

	e := c.Parse([]byte(`127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "INFO" {
		t.Errorf("Level = %q, want INFO", e.Level)
	}
	if e.Message != "GET /apache_pb.gif" {
		t.Errorf("Message = %q, want 'GET /apache_pb.gif'", e.Message)
	}
	if e.Fields["host"] != "127.0.0.1" {
		t.Errorf("Fields[host] = %v, want 127.0.0.1", e.Fields["host"])
	}
	if e.Fields["status"] != 200 {
		t.Errorf("Fields[status] = %v, want 200", e.Fields["status"])
	}
	if e.Fields["size"] != 2326 {
		t.Errorf("Fields[size] = %v, want 2326", e.Fields["size"])
	}
}

func TestCLFParseLevels(t *testing.T) {
	c := CLF{}

	cases := []struct {
		status string
		level  string
	}{
		{"200", "INFO"},
		{"301", "INFO"},
		{"404", "WARN"},
		{"500", "ERROR"},
		{"503", "ERROR"},
	}

	for _, tc := range cases {
		line := `10.0.0.1 - - [12/Apr/2025:14:23:05 +0000] "GET / HTTP/1.1" ` + tc.status + ` 100`
		e := c.Parse([]byte(line))
		if e.Level != tc.level {
			t.Errorf("status %s: Level = %q, want %q", tc.status, e.Level, tc.level)
		}
	}
}

func TestCLFParseCombined(t *testing.T) {
	c := CLF{}

	e := c.Parse([]byte(`10.0.0.1 - - [12/Apr/2025:14:23:05 +0000] "GET /health HTTP/1.1" 200 18 "https://example.com/start" "Mozilla/5.0 (compatible)"`))
	if e.Fields["referrer"] != "https://example.com/start" {
		t.Errorf("Fields[referrer] = %v, want https://example.com/start", e.Fields["referrer"])
	}
	if e.Fields["agent"] != "Mozilla/5.0 (compatible)" {
		t.Errorf("Fields[agent] = %v, want Mozilla/5.0 (compatible)", e.Fields["agent"])
	}
}
