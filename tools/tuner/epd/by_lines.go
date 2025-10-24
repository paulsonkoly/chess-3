package epd

import (
	"bufio"
	"io"
	"os"
)

// ByLines is a sequential reader for a named file, reading lines not including
// '\n', and it avoids allocations by yielding the underlying buffer slices.
type ByLines struct {
	f *os.File
	b *bufio.Reader
}

// OpenByLines opens the named file fn, and returns a ByLines reader.
func OpenByLines(fn string) (*ByLines, error) {
	iof, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	return &ByLines{f: iof, b: bufio.NewReader(iof)}, nil
}

// Close closes b.
func (b *ByLines) Close() error {
	return b.f.Close()
}

// Read reads a single line not including '\n' and skipping over empty lines
// from b.
func (b *ByLines) Read() ([]byte, error) {
	for { // skip over empties
		bytes, err := b.b.ReadSlice('\n')
		if err != nil {
			return nil, err
		}

		if len(bytes) > 1 {
			return bytes[:len(bytes)-1], nil // remove the '\n'
		}
	}
}

// Rewind resets the internal state of b, new read will start at the 0 file
// offset.
func (b *ByLines) Rewind() error {
	if _, err := b.f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	b.b.Reset(b.f)
	return nil
}
