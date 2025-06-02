package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/eval"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

var epdF = flag.String("epd", "", "epd file name")
var misEval = flag.Bool("misEval", false, "print top 10 misevaluated positions")
var filter = flag.Bool("filter", false, "filter out non-quiet or terminal node entries")
var cpuProf = flag.String("cpuProf", "", "cpu profile file name")
var diff = flag.String("diff", "", "output positions from this file not present in epd (result label ignored)")

// var memProf = flag.String("memProf", "", "mem profile file name")

type EPDEntry struct {
	b *board.Board
	r float64
}

const (
	BatchSize  = 100_000
	Epsilon    = 0.001
	NumThreads = 4
)

func main() {
	flag.Parse()

	if *cpuProf != "" {
		cpu, err := os.Create(*cpuProf)
		if err != nil {
			panic(err)
		}
		err = pprof.StartCPUProfile(cpu)
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	if *epdF == "" {
		panic("epd file name is required")
	}

	// read in the data set
	data := make([]EPDEntry, 0)
	f, err := os.Open(*epdF)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scn := bufio.NewScanner(f)
	cnt := 0
	for scn.Scan() {
		line := scn.Text()

		splits := strings.Split(line, "; ")

		if len(splits) != 2 {
			panic("epd line error " + line)
		}

		b := Must(board.FromFEN(splits[0]))
		// fmt.Println(b.FEN())
		r, err := strconv.ParseFloat(splits[1], 64)
		if err != nil {
			panic(err)
		}

		data = append(data, EPDEntry{b, r})
		cnt++
	}

	if *diff != "" {
		doDiff(data, diff)
		return
	}

	if *filter {
		doFilter(data)
		return
	}

	fmt.Println("initial coefficients:")
	coeffs := tuning.InitialCoeffs()
	coeffs.Print()

	k := 0.832 // a scaling constant

	// minimize for k
	improved := true
	step := 1.0
	bestE := computeE(data, k, coeffs)
	for step > 0.0001 {
		fmt.Println("step", step)
		for improved {
			eHigh := computeE(data, k+step, coeffs)
			eLow := computeE(data, k-step, coeffs)
			improved = false

			nK := k - step
			nE := eLow
			if eHigh < eLow {
				nK = k + step
				nE = eHigh
			}

			if nE < bestE {
				improved = true
				bestE = nE
				k = nK
				fmt.Println("new k value: ", k)
			}
		}
		step /= 10.0
		improved = true
	}

	if *misEval {
		printMisEval(data, k, coeffs)
		return
	}

	fmt.Println("bestE ", bestE)

	momentum := tuning.Coeffs{}
	velocity := tuning.Coeffs{}

	beta1 := 0.9
	beta2 := 0.999
	learningRate := 1.0

	bStart := 0
	for epoch := 1; true; { // epochs
		batch := data[bStart:min(bStart+BatchSize, len(data))]

		// calculate the gradients
		grad := tuning.Coeffs{}
		gradS := [NumThreads]tuning.Coeffs{}
		wg := sync.WaitGroup{}
		wg.Add(NumThreads)
		tSize := len(batch) / NumThreads

		for ix := range NumThreads {
			tData := batch[ix*tSize : min((ix+1)*tSize, len(batch))]
			coeffsS := *coeffs

			go func() {
				calcGradients(tData, &gradS[ix], &coeffsS, k)
				wg.Done()
			}()
		}

		wg.Wait()

		for ix := range NumThreads {
			for ixs := range grad.Loop() {
				*grad.At(ixs) += *gradS[ix].At(ixs)
			}
		}

		for ixs := range coeffs.Loop() {
			g := *grad.At(ixs) / float64(len(batch))
			m := momentum.At(ixs)
			*m = beta1*(*m) + (1-beta1)*g
			v := velocity.At(ixs)
			*v = beta2*(*v) + (1-beta2)*math.Pow(g, 2)

			mHat := *m / (1 - math.Pow(beta1, float64(epoch)))
			vHat := *v / (1 - math.Pow(beta2, float64(epoch)))

			c := coeffs.At(ixs)
			*c -= learningRate * mHat / (1e-8 + math.Sqrt(vHat))
		}

		bStart += BatchSize
		if bStart >= len(data) {
			fmt.Printf("epoch %d finished\n", epoch)

			newE := computeE(data, k, coeffs)
			fmt.Printf("error drop %.10f , bestE %.10f\n", bestE-newE, newE)
			if newE > bestE {
				fmt.Printf("drop negative, LR %.4f -> %.4f\n", learningRate, learningRate/2.0)
				learningRate /= 2
			}
			bestE = newE

			coeffs.Print()

			rand.Shuffle(len(data), func(i, j int) {
				data[i], data[j] = data[j], data[i]
			})

			epoch++
			bStart = 0
		}
	}
}

func calcGradients(data []EPDEntry, grad *tuning.Coeffs, coeffs *tuning.Coeffs, k float64) {
	for _, e := range data {
		score := evalCoeffs(e.b, coeffs)
		sigm := sigmoid(score, k)
		loss := (e.r - sigm) * (e.r - sigm)
		for ixs := range grad.Loop() {
			c := coeffs.At(ixs)
			o := *c

			*c += Epsilon
			score2 := evalCoeffs(e.b, coeffs)
			*c = o

			sigm2 := sigmoid(score2, k)
			loss2 := (e.r - sigm2) * (e.r - sigm2)

			g := (loss2 - loss) / Epsilon

			*grad.At(ixs) += g
		}
	}
}

func computeE(data []EPDEntry, k float64, coeffs *tuning.Coeffs) float64 {
	sum := 0.0
	count := 0

	for _, epdE := range data {
		b := epdE.b
		r := epdE.r

		score := evalCoeffs(b, coeffs)

		sgm := sigmoid(score, k)

		sum += (r - sgm) * (r - sgm)

		count++
	}

	return sum / float64(count)
}

func evalCoeffs(b *board.Board, coeffs *tuning.Coeffs) float64 {
	score := eval.Eval(b, coeffs.ToEvalType())

	if b.STM == Black {
		score = -score
	}
	return score
}

func sigmoid(v, k float64) float64 {
	return 1 / (1 + math.Exp(-k*v/400))
}

func printMisEval(data []EPDEntry, k float64, coeffs *tuning.Coeffs) {
	evals := make([]float64, len(data))

	for i, epdE := range data {
		b := epdE.b
		r := epdE.r

		score := evalCoeffs(b, coeffs)

		sgm := sigmoid(score, k)

		err := math.Abs(r - sgm)
		evals[i] = err
	}

	for range 10 {
		mx := math.Inf(-1)
		mi := -1
		for i, e := range evals {
			if e > mx {
				mx = e
				mi = i
			}
		}

		fmt.Printf("%s error %f.4 result %.1f eval %4f\n", data[mi].b.FEN(), mx, data[mi].r, evalCoeffs(data[mi].b, coeffs))
		evals[mi] = math.Inf(-1)
	}
}

func doFilter(data []EPDEntry) {
	ms := move.NewStore()

	for _, epdE := range data {
		b := epdE.b

		// if we are in check it's not a quiet position
		if movegen.InCheck(b, b.STM) {
			continue
		}

		movegen.GenMoves(ms, b)

		hasLegal := false
		hasForcing := false

		for _, m := range ms.Frame() {
			b.MakeMove(&m)

			if movegen.InCheck(b, b.STM.Flip()) { // move not legal
				b.UndoMove(&m)
				continue
			}
			hasLegal = true

			if m.Captured != NoPiece {
				b.UndoMove(&m)
				hasForcing = true
				break
			}

			if movegen.InCheck(b, b.STM) { // move is check, therefore forcing
				b.UndoMove(&m)
				hasForcing = true
				break
			}
			b.UndoMove(&m)
		}

		if hasLegal && !hasForcing {
			fmt.Printf("%s; %.1f\n", b.FEN(), epdE.r)
		}

		ms.Clear()
	}
}

func doDiff(data []EPDEntry, diff *string) {
	hsh := make(map[board.Hash]struct{})

	for _, d := range data {
		hsh[d.b.Hash()] = struct{}{}
	}

	f, err := os.Open(*diff)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scn := bufio.NewScanner(f)
	for scn.Scan() {
		line := scn.Text()

		splits := strings.Split(line, "; ")

		if len(splits) != 2 {
			panic("epd line error " + line)
		}

		b := Must(board.FromFEN(splits[0]))

		bHsh := b.Hash()
		if _, ok := hsh[bHsh]; !ok {
			fmt.Println(line)
		}
	}
}
