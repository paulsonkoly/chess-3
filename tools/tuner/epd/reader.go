package epd

import (
	"bufio"
	"errors"
	"io"
	"math/bits"
	"os"
	"slices"

	"golang.org/x/sys/unix"
)

const PagesPerMap = 32

// Reader provides a file reader, that reads in chunks from a file through some
// random shuffle determined by the epoch number.
//
// # Example
//
// r, _ := Open("file.epd")
// defer r.Close()
//
// // create a mapping for a chunk from line 10 to 30 through the shuffle for epoch 7.
// map := r.Map(7, 10, 30)
//
//	for {
//	  line, err := map.Read()
//	  if err != nil {
//	    if err == io.EOF {
//	      break
//	    }
//	    panic(err)
//	  }
//	  fmt.Println(line) // line is from line 10 to 30 of the file, as if the file was shuffled for epoch 7
//	}
type Reader struct {
	f            *os.File
	lineManifest []lineAddr
}

func (r Reader) LineCount() int { return len(r.lineManifest) }

type lineAddr struct {
	start int64
	end   int64
}

func Open(fn string) (*Reader, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	result := Reader{f: f, lineManifest: make([]lineAddr, 0)}

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
				result.lineManifest = append(result.lineManifest, lineAddr{start: prev, end: count})
			}
			prev = count + 1
		}
		count++
	}

	return &result, nil
}

func (r *Reader) Close() error {
	r.lineManifest = nil
	return r.f.Close()
}

type Map struct {
	chunkLines       []lineAddr
	chunkLinesIx     int
	f                *os.File
	mapStart, mapEnd int64
	mapBytes         []byte
}

var ErrClosedFile = errors.New("file is closed")
var ErrChunkInvalid = errors.New("invalid chunk")

func (r *Reader) Map(epoch, start, end int) (*Map, error) {
	if r.lineManifest == nil {
		return nil, ErrClosedFile
	}

	if start < 0 || end < 0 || start > len(r.lineManifest)-1 || end > len(r.lineManifest) || start > end {
		return nil, ErrChunkInvalid
	}

	chunkLines := make([]lineAddr, 0, end-start)
	for ix := start; ix < end; ix++ {
		line := r.lineManifest[shuffleIndex(uint64(ix), uint64(len(r.lineManifest)), uint64(epoch))]
		chunkLines = append(chunkLines, line)
	}

	// sort chunkLines so we can read at a reasonable rate. We are accessing at
	// random non-consecutive locations, but at least in the order of the
	// physical file.
	slices.SortFunc(chunkLines, func(a, b lineAddr) int {
		return int(a.start - b.start)
	})

	return &Map{chunkLines: chunkLines, f: r.f}, nil
}

func (m *Map) Read() (string, error) {
	if m.chunkLinesIx < 0 || m.chunkLinesIx >= len(m.chunkLines) {
		if m.mapBytes != nil {
			err := unix.Munmap(m.mapBytes)
			if err != nil {
				return "", err
			}
		}
		return "", io.EOF
	}

	addr := m.chunkLines[m.chunkLinesIx]

	mapSize := PagesPerMap * int64(unix.Getpagesize())

	startPIx := addr.start / mapSize
	endPIx := addr.end / mapSize

	reqMapStart := startPIx * mapSize
	reqMapEnd := (endPIx+1)*mapSize - 1

	if reqMapStart < m.mapStart || reqMapEnd > m.mapEnd {
		if m.mapBytes != nil {
			err := unix.Munmap(m.mapBytes)
			if err != nil {
				return "", err
			}
		}

		var err error
		m.mapBytes, err = unix.Mmap(
			int(m.f.Fd()),
			startPIx*mapSize,
			int((endPIx-startPIx+1)*mapSize),
			unix.PROT_READ,
			unix.MAP_SHARED)
		if err != nil {
			return "", err
		}
		m.mapStart = reqMapStart
		m.mapEnd = reqMapEnd
	}
	m.chunkLinesIx++
	return string(m.mapBytes[addr.start-m.mapStart : addr.end-m.mapStart]), nil
}

// Reset resets reading from a mapping.
func (m *Map) Reset() error {
	m.chunkLinesIx = 0
	m.mapStart = 0
	m.mapEnd = 0
	if m.mapBytes != nil {
		err := unix.Munmap(m.mapBytes)
		if err != nil {
			return err
		}
	}
	m.mapBytes = nil
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
