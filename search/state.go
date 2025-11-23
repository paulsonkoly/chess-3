package search

import (
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/transp"

	. "github.com/paulsonkoly/chess-3/types"
)

// Search contains the permanent stores such as tt that can be re-used between
// searches.
type Search struct {
	tt     *transp.Table
	hist   *heur.History
	cont   [2]*heur.Continuation
	ms     *move.Store
	hstack *historyStack
	pv     *pv
	gen    transp.Gen
}

// New creates a new Search object.
func New(size int) *Search {
	return &Search{
		tt:     transp.New(size),
		ms:     move.NewStore(),
		hist:   heur.NewHistory(),
		cont:   [2]*heur.Continuation{heur.NewContinuation(), heur.NewContinuation()},
		hstack: newHistStack(),
		pv:     newPV(),
	}
}

// refresh prepares the state for a new search.
func (s *Search) refresh() {
	s.ms.Clear()
	s.hstack.reset()
	s.pv.setNull(0)
}

// Clear clears the internal stores in the Search object. Should be called between games only.
func (s *Search) Clear() {
	s.gen = 0
	s.tt.Clear()
	s.hist.Clear()
	s.cont[0].Clear()
	s.cont[1].Clear()
}

type options struct {
	stop     chan struct{}
	abort    bool
	debug    bool
	softTime int64
	counters *Counters
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

// WithCounters instructs the search to collect statistics in counters.
func WithCounters(counters *Counters) Option {
	return func(o *options) {
		o.counters = counters
	}
}

// Soft time limit in milliseconds. <= 0 for no limit.
// TODO: should this be time.Duration?
func WithSoftTime(st int64) Option {
	return func(o *options) {
		o.softTime = st
	}
}

// Counters are various search counters.
type Counters struct {
	AWFail int // AwFail is the count of times the score fell outside of the aspiration window.
	ABLeaf int // ABLeaf is the count of alpha-beta leafs.
	// ABBreadth is the total count of explored moves in alpha-beta. Thus
	// (ABBreadth / ABCnt) is the average alpha-beta branching factor.
	ABBreadth int
	ABCnt     int   // ABCnt is the inner node count in alpha-beta.
	TTHit     int   // TThit is the transposition table hit-count.
	QCnt      int   // Quiesence node count
	QDepth    Depth // QDepth is the maximal quiesence search depth.
	Time      int64 // Time is the search time in milliseconds.
}
