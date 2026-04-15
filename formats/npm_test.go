package formats

import "testing"

func TestNpmProbe(t *testing.T) {
	n := Npm{}

	if !n.Probe([]byte(`0 verbose cli C:\Users\agustin.rodriguez\AppData\Local\Programs\runjs\RunJS.exe`)) {
		t.Error("expected true for verbose line")
	}
	if !n.Probe([]byte(`1 info using npm@10.8.2`)) {
		t.Error("expected true for info line")
	}
	if !n.Probe([]byte(`3 silly config load:file:C:\Users\agustin.rodriguez\.npmrc`)) {
		t.Error("expected true for silly line")
	}
	if !n.Probe([]byte(`42 warn deprecated package@1.0.0`)) {
		t.Error("expected true for warn line")
	}

	// plain text without seq number
	if n.Probe([]byte(`info using npm@10.8.2`)) {
		t.Error("expected false for missing seq number")
	}

	// JSON
	if n.Probe([]byte(`{"level":"info","msg":"hi"}`)) {
		t.Error("expected false for JSON")
	}

	// unknown level word
	if n.Probe([]byte(`5 blah something happened`)) {
		t.Error("expected false for unknown level")
	}
}

func TestNpmParse(t *testing.T) {
	n := Npm{}

	// verbose → DEBUG
	e := n.Parse([]byte(`0 verbose cli /usr/bin/node`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "DEBUG" {
		t.Errorf("verbose: got %q, want DEBUG", e.Level)
	}
	if e.Message != "cli /usr/bin/node" {
		t.Errorf("msg = %q", e.Message)
	}

	// silly → DEBUG
	e2 := n.Parse([]byte(`3 silly config load:file:/home/user/.npmrc`))
	if e2.Level != "DEBUG" {
		t.Errorf("silly: got %q, want DEBUG", e2.Level)
	}

	// info → INFO
	e3 := n.Parse([]byte(`1 info using npm@10.8.2`))
	if e3.Level != "INFO" {
		t.Errorf("info: got %q, want INFO", e3.Level)
	}

	// warn → WARN
	e4 := n.Parse([]byte(`42 warn deprecated package@1.0.0: use newpackage instead`))
	if e4.Level != "WARN" {
		t.Errorf("warn: got %q, want WARN", e4.Level)
	}

	// error → ERROR
	e5 := n.Parse([]byte(`99 error code ENOENT`))
	if e5.Level != "ERROR" {
		t.Errorf("error: got %q, want ERROR", e5.Level)
	}

	// timing → DEBUG
	e6 := n.Parse([]byte(`7 timing npm:load COMPLETEDIN 42ms`))
	if e6.Level != "DEBUG" {
		t.Errorf("timing: got %q, want DEBUG", e6.Level)
	}

	// no timestamp — zero value is fine
	if !e.Timestamp.IsZero() {
		t.Error("npm logs have no timestamp — should be zero")
	}
}
