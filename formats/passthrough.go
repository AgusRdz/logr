package formats

// Passthrough is the last-resort format for non-JSON lines.
type Passthrough struct{}

func (Passthrough) Probe(_ []byte) bool { return true }

func (Passthrough) Parse(line []byte) Entry {
	return Entry{
		ParseErr: true,
		Raw:      line,
		Message:  string(line),
	}
}
