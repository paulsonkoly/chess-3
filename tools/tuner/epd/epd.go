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
	"errors"
	"io"
	"math/bits"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/tools/tuner/checksum"
	"golang.org/x/sys/unix"
)

type lineAddr struct {
	start int64
	end   int64
}

type File struct {
	filename     string
	lineManifest []lineAddr
}

func (e File) Basename() string {
	return path.Base(e.filename)
}

func (e File) LineCount() int {
	return len(e.lineManifest)
}

// New creates an EPD file reader.
func New(filename string) (*File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	e := File{filename: filename, lineManifest: make([]lineAddr, 0)}

	scn := bufio.NewReader(f)
	count := int64(0)
	prev := int64(0)
	for {
		b, err := scn.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return nil, err
			}
		}

		if b == '\n' {
			// skip empty lines from the manifest
			if count > prev {
				e.lineManifest = append(e.lineManifest, lineAddr{start: prev, end: count})
			}
			prev = count + 1
		}
		count++
	}

	return &e, nil
}

// Checksum is the sha256 checksum of the whole content of the epd file.
// Concurrency safe.
func (e *File) Checksum() (checksum.Checksum, error) {
	f, err := os.Open(e.filename)
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

// Stream streams all content from e on a per line basis. Concurrency safe.
func (e *File) Stream(s Streamer) error {

	f, err := os.Open(e.filename)
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

var ErrChunkInvalid = errors.New("invalid chunk")
var ErrPageSize = errors.New("invalid page size")
var ErrLineInvalid = errors.New("invalid epd line")

// Entry is a parsed entry from the EPD file.
// TODO: should this be here?
type Entry struct {
	Board  *board.Board // Board is a chess position structure.
	Result float64      // Result is the WDL label.
}

// Chunk returns the lines of an EPD chunk, identified by the starting and
// ending line indices within a given epoch (shuffle). As it usually happens in
// go; start is inclusive, end is non-inclusive. Concurrency safe.
func (epdF *File) Chunk(epoch, start, end int) ([]Entry, error) {
	e := make(entries, 0)

	err := epdF.chunkReader(epoch, start, end, &e)
	if err != nil {
		return nil, err
	}

	return e, err
}

// Chunk returns the checksum of an EPD chunk, identified by the starting and
// ending line indices within a given epoch (shuffle). As it usually happens in
// go; start is inclusive, end is non-inclusive. Concurrency safe.
func (epdF *File) ChunkChecksum(epoch, start, end int) (checksum.Checksum, error) {
	collector := checksum.NewCollector()

	err := epdF.chunkReader(epoch, start, end, &collector)
	if err != nil {
		return checksum.Checksum{}, err
	}

	return collector.Checksum(), nil
}

type entries []Entry

func (e *entries) Collect(line string) error {
	splits := strings.Split(line, "; ")
	if len(splits) != 2 {
		return ErrLineInvalid
	}

	b, err := board.FromFEN(splits[0])
	if err != nil {
		return err
	}

	r, err := strconv.ParseFloat(splits[1], 64)
	if err != nil {
		return err
	}

	*e = append(*e, Entry{Board: b, Result: r})
	return nil
}

type collector interface {
	Collect(line string) error
}

func (e *File) chunkReader(epoch, start, end int, c collector) error {
	if start < 0 || end < 0 || start > len(e.lineManifest)-1 || end > len(e.lineManifest) || start > end {
		return ErrChunkInvalid
	}

	f, err := os.Open(e.filename)
	if err != nil {
		return err
	}

	chunkLines := make([]lineAddr, 0)
	for ix := start; ix < end; ix++ {
		line := e.lineManifest[shuffleIndex(uint64(ix), uint64(len(e.lineManifest)), uint64(epoch))]
		chunkLines = append(chunkLines, line)
	}

	// sort chunkLines so we can read at a reasonable rate. We are accessing at
	// random non-consecutive locations, but at least in the order of the
	// physical file.
	slices.SortFunc(chunkLines, func(a, b lineAddr) int {
		return int(a.start - b.start)
	})

	pageSize := int64(unix.Getpagesize())

	if pageSize&(pageSize-1) != 0 {
		return ErrPageSize
	}

	mapStart := int64(0)
	mapEnd := int64(0)
	var mapBytes []byte

	for _, addr := range chunkLines {

		startPIx := addr.start / pageSize
		endPIx := addr.end / pageSize

		reqMapStart := startPIx * pageSize
		reqMapEnd := (endPIx+1)*pageSize - 1

		if reqMapStart < mapStart || reqMapEnd > mapEnd {
			if mapBytes != nil {
				err := unix.Munmap(mapBytes)
				if err != nil {
					return err
				}
			}

			mapBytes, err = unix.Mmap(
				int(f.Fd()),
				startPIx*pageSize,
				int((endPIx-startPIx+1)*pageSize),
				unix.PROT_READ,
				unix.MAP_SHARED)
			if err != nil {
				return err
			}
			mapStart = reqMapStart
			mapEnd = reqMapEnd
		}

		line := string(mapBytes[addr.start-mapStart : addr.end-mapStart])

		c.Collect(line)
	}

	if mapBytes != nil {
		err := unix.Munmap(mapBytes)
		if err != nil {
			return err
		}
	}

	return nil
}

// shuffleIndex returns a pseudo-random permutation of x in [0, n)
// determined by the given seed. It’s a Feistel network, bijective for any n.
func shuffleIndex(x, n, seed uint64) uint64 {
	if n <= 1 {
		return 0
	}

	// Find smallest power of two ≥ n
	bitsNeeded := bits.Len64(n - 1)
	size := uint64(1) << bitsNeeded
	mask := size - 1

	for {
		y := feistel(x, seed, bitsNeeded)
		if y < n {
			return y
		}
		// Rejection sampling: try again with new x
		x = y & mask
	}
}

// Internal Feistel permutation on [0, 2^bits)
func feistel(x, seed uint64, bits int) uint64 {
	half := bits / 2
	leftMask := (uint64(1) << half) - 1
	rightMask := (uint64(1) << (bits - half)) - 1

	left := x & leftMask
	right := (x >> half) & rightMask

	const rounds = 4
	for i := range rounds {
		k := seed + uint64(i)*0x9e3779b97f4a7c15
		f := roundFunc(right, k) & leftMask
		left, right = right, left^f
	}

	return ((right & rightMask) << half) | (left & leftMask)
}

func roundFunc(x, k uint64) uint64 {
	z := x + k
	z ^= z >> 21
	z *= 0x9e3779b97f4a7c15
	z ^= z >> 33
	return z
}
