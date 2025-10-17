package tuning

import "iter"

// Range is an index range. Start is inclusive End is non-inclusive.
type Range struct {
	Start int
	End   int
}

func Batches(numEntries int) iter.Seq[Range] {
	return func(yield func(Range) bool) {
		for start := 0; start < numEntries; start += NumLinesInBatch {
			end := min(start+NumLinesInBatch, numEntries)
			if !yield(Range{Start: start, End: min(numEntries, end)}) {
				return
			}
		}
	}
}

func Chunks(batch Range) iter.Seq[Range] {
	return func(yield func(Range) bool) {
		numLinesInChunk := (NumLinesInBatch + NumChunksInBatch - 1) / NumChunksInBatch
		for start := batch.Start; start < batch.End; start += numLinesInChunk {
			end := min(start+numLinesInChunk, batch.End)
			if !yield(Range{Start: start, End: end}) {
				return
			}
		}
	}
}
