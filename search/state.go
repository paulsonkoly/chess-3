package search

import (
	"github.com/paulsonkoly/chess-3/heur"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/transp"

	. "github.com/paulsonkoly/chess-3/types"
)

type Search struct {
	tt     *transp.Table
	hist   *heur.History
	cont   [2]*heur.Continuation
	ms     *move.Store
	hstack *historyStack
	pv     *pv
}

func New(ttSizeInMb int) *Search {
	return &Search{
		tt:     transp.New(ttSizeInMb),
		ms:     move.NewStore(),
		hist:   heur.NewHistory(),
		cont:   [2]*heur.Continuation{heur.NewContinuation(), heur.NewContinuation()},
		hstack: newHistStack(),
		pv:     newPV(),
	}
}

func (s *Search) Clear() {
	s.tt.Clear()
	s.ms.Clear()
	s.hstack.reset()
}

type options struct {
	stop     chan struct{}
	abort    bool
	debug    bool
	softTime int64 
	counters *Counters
}

type Option = func(*options)

func WithStop(stop chan struct{}) Option {
	return func(o *options) {
		o.stop = stop
	}
}

func WithDebug(debug bool) Option {
	return func(o *options) {
		o.debug = debug
	}
}

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
