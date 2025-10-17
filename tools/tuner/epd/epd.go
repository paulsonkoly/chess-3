package epd

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"hash"
	"io"
	"math/rand/v2"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/paulsonkoly/chess-3/board"
	"golang.org/x/sys/unix"
)

type lineAddr struct {
	start int64
	end   int64
}

type File struct {
	filename     string
	f            *os.File
	lineManifest []lineAddr
	checksum     []byte
}

func (e File) Basename() string {
	return path.Base(e.filename)
}

func (e File) LineCount() int {
	return len(e.lineManifest)
}

// Open opens an EPD file.
func Open(filename string) (*File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	e := File{f: f, filename: filename, lineManifest: make([]lineAddr, 0)}

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

// Close closes the epd file and frees up associated resources.
func (e *File) Close() {
	e.f.Close()
	// free the lineManifest it can be massive
	e.lineManifest = nil
}

// Checksum is the sha256 checksum of the whole content of the epd file.
func (e *File) Checksum() ([]byte, error) {
	if e.checksum != nil {
		return e.checksum, nil
	}

	fd, err := unix.Dup(int(e.f.Fd()))
	if err != nil {
		return nil, err
	}

	f := os.NewFile(uintptr(fd), e.filename)
	defer f.Close()

	f.Seek(0, io.SeekStart)

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	e.checksum = h.Sum(nil)

	return e.checksum, nil
}

// Streamer represents a stream object for the lines of the file. Probably
// something that writes into a grpc stream.
type Streamer interface {
	Send(line string) error
}

// Stream streams all content from e on a per line basis.
func (e *File) Stream(s Streamer) error {
	fd, err := unix.Dup(int(e.f.Fd()))
	if err != nil {
		return err
	}

	f := os.NewFile(uintptr(fd), e.filename)
	defer f.Close()

	f.Seek(0, io.SeekStart)

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
type Entry struct {
	Board  *board.Board // Board is a chess position structure.
	Result float64      // Result is the WDL label.
}

// Chunk returns the lines of an EPD chunk, identified by the starting and
// ending line indices. As it usually happens in go; start is inclusive, end is
// non-inclusive.
func (epdF *File) Chunk(start, end int) ([]Entry, error) {
	e := make(entries, 0)

	err := epdF.chunkReader(start, end, e)
	if err != nil {
		return nil, err
	}

	return e, err
}

// ChunkChecksum returns the sha256 checksum of the given chunk.
func (epdF *File) ChunkChecksum(start, end int) ([]byte, error) {
	ch := checksum{sha256.New()}

	err := epdF.chunkReader(start, end, ch)
	if err != nil {
		return nil, err
	}

	return ch.h.Sum(nil), nil
}

type entries []Entry

func (e entries) Collect(line string) error {
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

	e = append(e, Entry{Board: b, Result: r})
	return nil
}

type checksum struct {
	h hash.Hash
}

func (c checksum) Collect(line string) error {
	c.h.Write([]byte(line))
	return nil
}

type collector interface {
	Collect(line string) error
}

func (e *File) chunkReader(start, end int, c collector) error {
	if start < 0 || end < 0 || start > len(e.lineManifest)-1 || end > len(e.lineManifest) || start > end {
		return ErrChunkInvalid
	}

	var fd int
	fd, err := unix.Dup(int(e.f.Fd()))
	if err != nil {
		return err
	}

	f := os.NewFile(uintptr(fd), e.filename)
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	chunkLines := make([]lineAddr, 0)
	for _, addr := range e.lineManifest[start:end] {
		chunkLines = append(chunkLines, addr)
	}

	// sort the chunk line manifest so we can read at a reasonable rate. We are
	// accessing at random non-consecutive locations, but in the order of the
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

// Shuffle shuffles the line order in the file, the order is determined by seed.
func (e *File) Shuffle(seed int) {
	src := rand.NewPCG(uint64(seed), uint64(seed)^uint64(0x9e3779b97f4a7c15))
	r := rand.New(src)

	// Fisher–Yates shuffle
	for i := len(e.lineManifest) - 1; i > 0; i-- {
		j := r.IntN(i + 1)
		e.lineManifest[i], e.lineManifest[j] = e.lineManifest[j], e.lineManifest[i]
	}
}
