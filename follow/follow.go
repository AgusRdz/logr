package follow

import (
	"bufio"
	"os"
	"time"
)

// Follow tails path, sending raw lines to out.
// Stops when done is closed.
// Uses os.SameFile for rotation detection.
// Seeks to end on first open (does not replay existing content).
// Polls every 100ms when at EOF.
func Follow(path string, out chan<- []byte, done <-chan struct{}) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	if _, err := f.Seek(0, 2); err != nil { // 2 = io.SeekEnd, avoid io import
		return
	}

	origInfo, err := f.Stat()
	if err != nil {
		return
	}

	reader := bufio.NewReaderSize(f, 64*1024)

	for {
		select {
		case <-done:
			return
		default:
		}

		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			// Strip trailing newline
			trimmed := line
			if n := len(trimmed); n > 0 && trimmed[n-1] == '\n' {
				trimmed = trimmed[:n-1]
			}
			if n := len(trimmed); n > 0 && trimmed[n-1] == '\r' {
				trimmed = trimmed[:n-1]
			}
			cp := make([]byte, len(trimmed))
			copy(cp, trimmed)
			select {
			case <-done:
				return
			case out <- cp:
			}
		}

		if err != nil {
			// EOF - poll
			select {
			case <-done:
				return
			case <-time.After(100 * time.Millisecond):
			}

			newInfo, statErr := os.Stat(path)
			if statErr != nil {
				continue
			}

			if !os.SameFile(origInfo, newInfo) {
				// File rotated - close old, reopen from start
				f.Close()
				f, err = os.Open(path)
				if err != nil {
					return
				}
				origInfo = newInfo
				reader = bufio.NewReaderSize(f, 64*1024)
			}
			// else: same file, reader will continue from current position
		}
	}
}
