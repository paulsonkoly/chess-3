package search

import (
	"io"
	"time"

	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/stack"
	"github.com/paulsonkoly/chess-3/transp"

	. "github.com/paulsonkoly/chess-3/chess"
)

// Search contains the permanent stores such as tt that can be re-used between
// searches.
type Search struct {
	tt      *transp.Table
	ranker  heur.MoveRanker
	ms      *move.Store
	hstack  *stack.Stack[heur.StackMove]
	pv      *pv
	gen     transp.Gen
	aborted bool
}

// New creates a new Search object.
func New(size int) *Search {
	return &Search{
		tt:     transp.New(size),
		ms:     move.NewStore(),
		ranker: heur.NewMoveRanker(),
		hstack: stack.New[heur.StackMove](),
		pv:     newPV(),
	}
}

// ResizeTT resizes the current tt to new size potentially re-allocating it.
func (s *Search) ResizeTT(size int) { s.tt.Resize(size) }

// refresh prepares the state for a new search.
func (s *Search) refresh() {
	s.ms.Clear()
	s.hstack.Reset()
	s.aborted = false
}

// Clear clears the internal stores in the Search object. Should be called between games only.
func (s *Search) Clear() {
	s.gen = 0
	s.tt.Clear()
	s.ranker.Clear()
}

// Options structure contains all search options. The functional options
// pattern can be used to invoke Go.Search, this struct does not have to be
// used directly.
type Options struct {
	Stop      chan struct{} // Stop channel interrupts the search.
	// Ponderhit channel signals a ponderhit. The sent time should be the time the ponderhit happend.
	PonderHit <-chan time.Time
	Output    io.Writer     // Info line output. nil for no output.
	SoftTime  int64         // SoftTime sets the soft timeout.
	Nodes     int           // Nodes sets the hard node count limit.
	SoftNodes int           // SoftNodes sets the soft node count limit.
	Counters  *Counters     // Counters sets the location where search has to gather statistics.
	Depth     Depth         // Depth sets the search depth limit.
	Debug     bool          // Debug turns extra debugging on.
}

// softAbort determines if elapsed times or nodes count justify a soft abort;
// that is aborting after a full completion of a given depth. Limits are
// ignored while pondering.
func (o *Options) softAbort(elapsed int64, nodes int) bool {
	return o.PonderHit == nil && ((o.SoftTime > 0 && elapsed > o.SoftTime) || (o.SoftNodes > 0 && nodes > o.SoftNodes))
}

// Option modifies how a search runs, this should be set per search.
type Option = func(*Options)

// WithStop runs the search with a stop channel. When the channel is signalled
// the search stops.
func WithStop(stop <-chan struct{}) Option {
	return func(o *Options) {
		o.Stop = stop
	}
}

// WithPonderHit runs the search with pondering. A signal on this channel
// indicates that the search should transition to normal search from a ponder
// search.
func WithPonderHit(ponderhit <-chan time.Time) Option {
	return func(o *Options) {
		o.PonderHit = ponderhit
	}
}

// WithDebug runs the search with debug outputs.
func WithDebug(debug bool) Option {
	return func(o *Options) {
		o.Debug = debug
	}
}

// WithOutput runs a search with outputs written to out. Replaces the default os.Stdout.
func WithOutput(out io.Writer) Option {
	return func(o *Options) {
		o.Output = out
	}
}

// WithCounters instructs the search to collect statistics in counters.
func WithCounters(counters *Counters) Option {
	return func(o *Options) {
		o.Counters = counters
	}
}

// WithSoftTime controls time limit in milliseconds. <= 0 for no limit.
// TODO: should this be time.Duration?
func WithSoftTime(st int64) Option {
	return func(o *Options) {
		o.SoftTime = st
	}
}

// WithDepth runs the search with depth limit. Useful for "go depth" uci command.
func WithDepth(d Depth) Option {
	return func(o *Options) { o.Depth = d }
}

// WithNodes runs the search with hard node count limit. Useful for "go nodes"
// uci command.
func WithNodes(nodes int) Option {
	return func(o *Options) { o.Nodes = nodes }
}

// WithSoftNodes sets a soft node count limit. When exceeded after completing
// a depth, the search will stop. <= 0 for no limit.
func WithSoftNodes(nodes int) Option {
	return func(o *Options) { o.SoftNodes = nodes }
}

// Counters are various search counters.
type Counters struct {
	Nodes   int   // Nodes is the total node count.
	ABNodes int   // ABNodes is the total node count limited to alpha-beta not counting leafs (d==0).
	Time    int64 // Time is the search time in milliseconds.
	// Moves is the total explored (searched) move count. Moves / ABNodes ~ avg. branching factor
	// Only counted if debug is set.
	Moves    int
	FirstCut int // FirstCut counts how many times AB searched exactly 1 move. Only counted if debug is set.
}
