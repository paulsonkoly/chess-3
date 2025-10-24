// package epd is designed to read large and potentially shuffled EPD + WDL
// data sets.
//
// While working with epd files, one must not modify the underlying physical
// file, including deleting, modifying data or any other file op.
package epd

import (
	"io"
	"os"

	"github.com/paulsonkoly/chess-3/tools/tuner/checksum"
)

// Checksum is the sha256 checksum of the whole content of a file.
func Checksum(fn string) (checksum.Checksum, error) {
	f, err := os.Open(fn)
	if err != nil {
		return checksum.Checksum{}, err
	}
	defer f.Close()
	return checksum.ReadFrom(f)
}

// Streamer represents a stream object for the lines of the file. Probably
// something that writes into a grpc stream.
type Streamer interface {
	Send(line string) error
}

// Stream streams all content from a file on a per line basis.
func Stream(fn string, s Streamer) error {
	byLines, err := OpenByLines(fn)
	if err != nil {
		return err
	}
	defer byLines.Close()

	for {
		bytes, err := byLines.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		s.Send(string(bytes))
	}
}
