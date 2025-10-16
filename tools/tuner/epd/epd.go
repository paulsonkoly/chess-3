package epd

import (
	"bufio"
	"crypto/sha256"
	"io"
	"iter"
	"math/rand/v2"
	"os"
	"path"

	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
)

type EPD struct {
	lines    []string
	filename string
	Checksum []byte
}

func Load(fn string) (*EPD, error) {
	lines := make([]string, 0)

	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scn := bufio.NewScanner(f)
	for scn.Scan() {
		lines = append(lines, scn.Text())
	}

	f.Seek(0, io.SeekStart)

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return &EPD{lines: lines, filename: fn, Checksum: h.Sum(nil)}, nil
}

func (e EPD) Basename() string { return path.Base(e.filename) }

func (e *EPD) Batches() iter.Seq[tuning.Batch] {
	return func(yield func(tuning.Batch) bool) {
		for start := 0; start < len(e.lines); start += tuning.NumLinesInBatch {
			b := tuning.Batch{Start: start, End: min(start+tuning.NumLinesInBatch-1, len(e.lines)-1)}

			if !yield(b) {
				return
			}
		}
	}
}

func (e *EPD) Shuffle() {
	rand.Shuffle(len(e.lines), func(i int, j int) {
		e.lines[i], e.lines[j] = e.lines[j], e.lines[i]
	})
}
