package epd

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"io"
	"math/rand/v2"
	"os"
	"slices"

	"golang.org/x/sys/unix"
)

type lineAddr struct {
	start int64
	end   int64
}

type EPD struct {
	filename     string
	f            *os.File
	lineManifest []lineAddr
	checksum     []byte
}

// Open opens an EPD file.
func Open(filename string) (*EPD, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	e := EPD{f: f, filename: filename, lineManifest: make([]lineAddr, 0)}

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
func (e *EPD) Close() {
	e.f.Close()
	// free the lineManifest it can be massive
	e.lineManifest = nil
}

// Checksum is the sha256 checksum of the whole content of the epd file.
func (e *EPD) Checksum() ([]byte, error) {
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

// Stream streams all content from e on a line basis.
func (e *EPD) Stream(s Streamer) error {
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

// Chunk returns the lines of an EPD chunk, identified by the starting and
// ending line indices. As it usually happens in go; start is inclusive, end is
// non-inclusive.
func (e *EPD) Chunk(start, end int) ([]string, error) {
	if start < 0 || end < 0 || start > len(e.lineManifest)-1 || end > len(e.lineManifest) || start > end {
		return nil, ErrChunkInvalid
	}

	var fd int
	fd, err := unix.Dup(int(e.f.Fd()))
	if err != nil {
		return nil, err
	}

	f := os.NewFile(uintptr(fd), e.filename)
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
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

	lines := make([]string, 0)
	pageSize := int64(unix.Getpagesize())

	if pageSize&(pageSize-1) != 0 {
		return nil, ErrPageSize
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
					return nil, err
				}
			}

			mapBytes, err = unix.Mmap(
				int(f.Fd()),
				startPIx*pageSize,
				int((endPIx-startPIx+1)*pageSize),
				unix.PROT_READ,
				unix.MAP_SHARED)
			if err != nil {
				return nil, err
			}
			mapStart = reqMapStart
			mapEnd = reqMapEnd
		}

		line := string(mapBytes[addr.start-mapStart : addr.end-mapStart])

		lines = append(lines, line)
	}

	if mapBytes != nil {
		err := unix.Munmap(mapBytes)
		if err != nil {
			return nil, err
		}
	}

	return lines, nil
}

// Shuffle shuffles the line order in the file, the order is determined by seed.
func (e *EPD) Shuffle(seed int) {
	src := rand.NewPCG(uint64(seed), uint64(seed)^uint64(0x9e3779b97f4a7c15))
	r := rand.New(src)

	// Fisher–Yates shuffle
	for i := len(e.lineManifest) - 1; i > 0; i-- {
		j := r.IntN(i + 1)
		e.lineManifest[i], e.lineManifest[j] = e.lineManifest[j], e.lineManifest[i]
	}
}
