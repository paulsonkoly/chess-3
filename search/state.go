package search

import (
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

type options struct {
	stop      chan struct{}
	softTime  int64
	nodes     int
	softNodes int
	counters  *Counters
	depth     Depth
	debug     bool
	info      bool
}

// softAbort determines if elapsed times or nodes count justify a soft abort;
// that is aborting after a full completion of a given depth.
func (o *options) softAbort(elapsed int64, nodes int) bool {
	return (o.softTime > 0 && elapsed > o.softTime) || (o.softNodes > 0 && nodes > o.softNodes)
}

// Option modifies how a search runs, this should be set per search.
type Option = func(*options)

// WithStop runs the search with a stop channel. When the channel is signalled
// the search stops.
func WithStop(stop chan struct{}) Option {
	return func(o *options) {
		o.stop = stop
	}
}

// WithDebug runs the search with debug outputs.
func WithDebug(debug bool) Option {
	return func(o *options) {
		o.debug = debug
	}
}

// WithInfo runs the search with uci info line outputs.
func WithInfo(info bool) Option {
	return func(o *options) {
		o.info = info
	}
}

// WithCounters instructs the search to collect statistics in counters.
func WithCounters(counters *Counters) Option {
	return func(o *options) {
		o.counters = counters
	}
}

// WithSoftTime controls time limit in milliseconds. <= 0 for no limit.
// TODO: should this be time.Duration?
func WithSoftTime(st int64) Option {
	return func(o *options) {
		o.softTime = st
	}
}

// WithDepth runs the search with depth limit. Useful for "go depth" uci command.
func WithDepth(d Depth) Option {
	return func(o *options) { o.depth = d }
}

// WithNodes runs the search with hard node count limit. Useful for "go nodes"
// uci command.
func WithNodes(nodes int) Option {
	return func(o *options) { o.nodes = nodes }
}

// WithSoftNodes sets a soft node count limit. When exceeded after completing
// a depth, the search will stop. <= 0 for no limit.
func WithSoftNodes(nodes int) Option {
	return func(o *options) { o.softNodes = nodes }
}

// Counters are various search counters.
type Counters struct {
	Nodes int   // Nodes is the total node count.
	Time  int64 // Time is the search time in milliseconds.
}
