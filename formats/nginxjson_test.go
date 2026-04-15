package formats

import "testing"

func TestNginxJSONProbe(t *testing.T) {
	n := NginxJSON{}

	if !n.Probe([]byte(`{"remote_addr":"10.0.0.1","time_local":"2025-04-12T14:23:05+00:00","request":"GET /api HTTP/1.1","status":"200"}`)) {
		t.Error("expected true for valid nginx JSON")
	}

	// missing remote_addr
	if n.Probe([]byte(`{"request":"GET /api HTTP/1.1","status":"200"}`)) {
		t.Error("expected false for missing remote_addr")
	}

	// missing request
	if n.Probe([]byte(`{"remote_addr":"10.0.0.1","status":"200"}`)) {
		t.Error("expected false for missing request")
	}

	// missing status
	if n.Probe([]byte(`{"remote_addr":"10.0.0.1","request":"GET /api HTTP/1.1"}`)) {
		t.Error("expected false for missing status")
	}

	// invalid JSON
	if n.Probe([]byte(`not json`)) {
		t.Error("expected false for invalid JSON")
	}
}

func TestNginxJSONParse(t *testing.T) {
	n := NginxJSON{}

	// 200 → INFO
	e := n.Parse([]byte(`{"remote_addr":"10.0.0.1","time_local":"2025-04-12T14:23:05+00:00","request":"GET /api/users HTTP/1.1","status":"200","body_bytes_sent":"1024","request_time":"0.042"}`))
	if e.ParseErr {
		t.Fatal("unexpected ParseErr")
	}
	if e.Level != "INFO" {
		t.Errorf("200: got %q, want INFO", e.Level)
	}
	if e.Message != "GET /api/users HTTP/1.1" {
		t.Errorf("msg = %q", e.Message)
	}

	// 404 → WARN
	e2 := n.Parse([]byte(`{"remote_addr":"10.0.0.2","time_local":"2025-04-12T14:23:06+00:00","request":"GET /missing HTTP/1.1","status":"404"}`))
	if e2.Level != "WARN" {
		t.Errorf("404: got %q, want WARN", e2.Level)
	}

	// 500 → ERROR
	e3 := n.Parse([]byte(`{"remote_addr":"10.0.0.3","time_local":"2025-04-12T14:23:07+00:00","request":"POST /api/pay HTTP/1.1","status":"500"}`))
	if e3.Level != "ERROR" {
		t.Errorf("500: got %q, want ERROR", e3.Level)
	}

	// numeric status
	e4 := n.Parse([]byte(`{"remote_addr":"10.0.0.4","time_local":"2025-04-12T14:23:08+00:00","request":"GET /api HTTP/1.1","status":503}`))
	if e4.Level != "ERROR" {
		t.Errorf("503 numeric: got %q, want ERROR", e4.Level)
	}

	// status and request removed from fields
	for _, k := range []string{"status", "request", "time_local"} {
		if _, ok := e.Fields[k]; ok {
			t.Errorf("field %q should be removed", k)
		}
	}

	// other fields kept
	if _, ok := e.Fields["body_bytes_sent"]; !ok {
		t.Error("body_bytes_sent should be in fields")
	}
}
