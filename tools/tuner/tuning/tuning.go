package tuning

const (
	ExitFailure = 1
	// NumLinesInBatch determines how the epd file is split into batches. A batch
	// completion implies the coefficients update.
	NumLinesInBatch = 100_000

	// NumChunksInBatch determines how a batch is split into chunks. A chunk is a
	// unique work iterm handed over to clients.
	NumChunksInBatch = 16
)

type Coeffs []float64

func (c *Coeffs) Add(other Coeffs) {
	for i := range *c {
		(*c)[i] += other[i]
	}
}
