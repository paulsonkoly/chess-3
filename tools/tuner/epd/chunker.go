package epd

import (
	"errors"
	"io"
	"math/bits"
	"os"
	"slices"
)

const (
	backingBytes = 32 * 1024 * 1024 // should give us 32Mb maps
)

var (
	ErrChunkInvalid  = errors.New("invalid chunk")
	ErrTruncatedRead = errors.New("truncated read")
)

// Chunker provides a file reader that reads in chunks from a file through some
// random shuffle determined by the epoch number.
//
// # Example
//
//	chunker, _ := NewChunker("file.epd")
//
//	// we want to read the section from 10 to 20 through the shuffle for epoch 7
//	chunk, _ := chunker.Open(7, 10, 20)
//	defer chunk.Close()
//
//	for {
//	 line, err := chunk.Read()
//	  if err != nil {
//	    if err == io.EOF {
//	      break
//	    }
//	    panic(err)
//	  }
//	  fmt.Println(line)
//	}
//
//	// we could re-use the same chunk here if we need to
//	chunk.Rewind()
type Chunker struct {
	fn           string
	lineManifest []lineAddr
}

type lineAddr struct {
	start int64
	end   int64 // non inclusive, and addresses the physical file which includes the '\n'
}

func NewChunker(fn string) (*Chunker, error) {
	byLines, err := OpenByLines(fn)
	if err != nil {
		return nil, err
	}
	defer byLines.Close()

	lineManifest := make([]lineAddr, 0)
	curr := int64(0)
	for {
		line, err := byLines.Read()
		if err != nil {
			if err == io.EOF {
				return &Chunker{fn: fn, lineManifest: lineManifest}, nil
			}
			return nil, err
		}
		end := curr + int64(len(line)) + 1 // +1 for '\n'
		lineManifest = append(lineManifest, lineAddr{curr, end})
		curr = end
	}
}

func (c Chunker) LineCount() int { return len(c.lineManifest) }

// Chunk is a virtual window to the physical file as if the order of the lines
// were shuffled based on the epoch. It limits reading between a start and end
// index. The indices are indexing in the shuffled order. Therefore a line
// might be returned that is outside of the start end range in the physical
// file order. Lines are read in the order of the physical file, potentially
// with gaps.
//
// For examples see Chunker.
type Chunk struct {
	f                *os.File
	chunkLines       []lineAddr
	chunkLinesIx     int
	mapStart, mapEnd int64
	mapBytes         []byte
}

// Open returns a new shuffled order window to the file.
func (c Chunker) Open(epoch, start, end int) (*Chunk, error) {
	if start < 0 || end < 0 || start > len(c.lineManifest)-1 || end > len(c.lineManifest) || start > end {
		return nil, ErrChunkInvalid
	}

	chunkLines := make([]lineAddr, 0, end-start)
	for ix := start; ix < end; ix++ {
		line := c.lineManifest[shuffleIndex(uint64(ix), uint64(len(c.lineManifest)), uint64(epoch))]
		chunkLines = append(chunkLines, line)
	}

	// sort chunkLines so we can read at a reasonable rate. We are accessing at
	// random non-consecutive locations, but at least in the order of the
	// physical file.
	slices.SortFunc(chunkLines, func(a, b lineAddr) int {
		return int(a.start - b.start)
	})

	mapBytes := make([]byte, backingBytes)

	f, err := os.Open(c.fn)
	if err != nil {
		return nil, err
	}

	return &Chunk{f: f, chunkLines: chunkLines, mapBytes: mapBytes}, nil
}

// Close closes the shuffled order window.
func (c *Chunk) Close() error {
	return c.f.Close()
}

// Read reads a line - not including '\n' from c.
func (c *Chunk) Read() ([]byte, error) {
	if c.chunkLinesIx < 0 || c.chunkLinesIx >= len(c.chunkLines) {
		return nil, io.EOF
	}

	// file relative addresses of what we need to read
	addr := c.chunkLines[c.chunkLinesIx]

	// is the line outside of the backing buffer
	if c.mapStart > addr.start || c.mapEnd < addr.end {
		cnt, err := c.f.ReadAt(c.mapBytes, addr.start)
		if err != nil && err != io.EOF {
			return nil, err
		}

		c.mapStart = addr.start
		c.mapEnd = addr.start + int64(cnt)
	}

	c.chunkLinesIx++
	return c.mapBytes[addr.start-c.mapStart : addr.end-c.mapStart-1], nil
}

// Reset resets reading from a mapping.
func (c *Chunk) Rewind() error {
	if _, err := c.f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	c.chunkLinesIx = 0
	c.mapStart = 0
	c.mapEnd = 0
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
