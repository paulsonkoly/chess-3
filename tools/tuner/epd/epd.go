// package epd is designed to read large and potentially shuffled EPD + WDL
// data sets.
//
// No portable way exists to keep a reusable FD open for concurrency-safe reads
// to the same inode. On Linux, /proc/self/fd/%d works, but Dup-ed FDs share
// read offsets and don’t solve concurrency. We accept potential errors if an
// EPD file is deleted or moved while open.
//
// Therefore one should not move / modify the underlying file while working
// with EPD.
package epd

import (
	"bufio"
	"os"

	"github.com/paulsonkoly/chess-3/tools/tuner/checksum"
)

// Checksum is the sha256 checksum of the whole content of a file.
// Concurrency safe.
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

// Stream streams all content from a file on a per line basis. Concurrency safe.
func Stream(fn string, s Streamer) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	scn := bufio.NewScanner(f)
	for scn.Scan() {
		err := s.Send(scn.Text())
		if err != nil {
			return err
		}
	}

	return nil
}
