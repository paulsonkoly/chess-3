package tuning

import "iter"

const (
	// NumLinesInBatch determines how the epd file is split into batches. A batch
	// completion implies the coefficients update.
	NumLinesInBatch = 100_000

	// NumChunksInBatch determines how a batch is split into chunks. A chunk is a
	// unique work iterm handed over to clients.
	NumChunksInBatch = 16
)

type Batch struct {
	Start int
	End   int
}

type Chunk struct {
	Start int
	End   int
}

func (b Batch) Chunks() iter.Seq[Chunk] {
	linesInChunk := (NumLinesInBatch + NumChunksInBatch - 1) / NumChunksInBatch

	return func(yield func(Chunk) bool) {
		for start := b.Start; start < b.End; start += linesInChunk {
			end := start + linesInChunk - 1
			if !yield(Chunk{Start: start, End: min(end, b.End)}) {
				return
			}
		}
	}
}

type Coeffs []float64

func (c *Coeffs) Add(other Coeffs) {
	for i := range *c {
		(*c)[i] += other[i]
	}
}
