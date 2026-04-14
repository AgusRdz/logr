package follow

import (
	"os"
	"testing"
	"time"
)

func TestFollow_ReadNewLines(t *testing.T) {
	f, err := os.CreateTemp("", "logr-follow-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	path := f.Name()

	out := make(chan []byte, 10)
	done := make(chan struct{})
	defer close(done)

	go Follow(path, out, done)

	// Give follow time to open and seek to end
	time.Sleep(300 * time.Millisecond)

	// Write lines after follow has started
	if _, err := f.WriteString("line one\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("line two\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	// Collect with timeout
	var got []string
	deadline := time.After(3 * time.Second)
collect:
	for len(got) < 2 {
		select {
		case line := <-out:
			got = append(got, string(line))
		case <-deadline:
			break collect
		}
	}

	if len(got) < 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(got), got)
	}
	if got[0] != "line one" {
		t.Errorf("expected 'line one', got %q", got[0])
	}
	if got[1] != "line two" {
		t.Errorf("expected 'line two', got %q", got[1])
	}
}

func TestFollow_Done(t *testing.T) {
	f, err := os.CreateTemp("", "logr-follow-done-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	out := make(chan []byte, 10)
	done := make(chan struct{})

	finished := make(chan struct{})
	go func() {
		Follow(f.Name(), out, done)
		close(finished)
	}()

	time.Sleep(50 * time.Millisecond)
	close(done)

	select {
	case <-finished:
		// good
	case <-time.After(2 * time.Second):
		t.Error("Follow did not stop after done was closed")
	}
}
