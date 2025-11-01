package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/google/uuid"
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/tools/tuner/app"
	"github.com/paulsonkoly/chess-3/tools/tuner/checksum"
	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	"github.com/paulsonkoly/chess-3/tools/tuner/shim"
	"github.com/paulsonkoly/chess-3/tools/tuner/tui"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
)

const (
	// JobTTLInit is the initial job time to live duration in seconds.
	JobTTLInit = 600
	// JobQueueDepth determines how many jobs the EPD processor can produce
	// before its stalled back.
	JobQueueDepth    = tuning.NumChunksInBatch / 2
	ResultQueueDepth = 10
	// ClientWaitTime is the duration in ms the server waits for results before
	// creating new jobs.
	ClientWaitTime = 100
)

func Run(args []string) {
	var epdFileName string
	var host string
	var port int
	var outFn string
	var minKPProf string
	var minKMProf string
	var useTui bool

	sFlags := flag.NewFlagSet("server", flag.ExitOnError)
	sFlags.StringVar(&epdFileName, "epd", "", "epd file name")
	sFlags.StringVar(&host, "host", "localhost", "host to listen on")
	sFlags.IntVar(&port, "port", 9001, "port to listen on")
	sFlags.StringVar(&outFn, "out", "coeffs.go", "coeff output file")
	sFlags.StringVar(&minKPProf, "kpprof", "", "filename for gathering cpu profiling data from the mse minimization with k")
	sFlags.StringVar(&minKMProf, "kmprof", "", "filename for gathering mem profiling data from the mse minimization with k")
	sFlags.BoolVar(&useTui, "tui", false, "render a tui")
	sFlags.Parse(args)

	if epdFileName == "" {
		slog.Error("no epd file given")
		os.Exit(app.ExitFailure)
	}

	ctx, cancel := context.WithCancel(context.Background())

	tuiQueue := make(chan tui.Update, tui.QueueDepth)

	go tui.Run(ctx, cancel, useTui, tuiQueue)

	k, err := minimizeK(ctx, epdFileName, tuiQueue, minKPProf, minKMProf)
	if err != nil {
		slog.Error("k minimization error", "error", err)
		os.Exit(app.ExitFailure)
	}

	jobQueue := make(chan shim.Job, JobQueueDepth)
	resultQueue := make(chan shim.Result, ResultQueueDepth)

	go epdProcess(ctx, epdFileName, outFn, k, jobQueue, resultQueue, tuiQueue)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		slog.Error("failed to bind port", "host", host, "port", port)
		os.Exit(app.ExitFailure)
	}

	tuiQueue <- tui.HostUpdate{Host: host, Port: port}
	srv := shim.NewServer(epdFileName, jobQueue, resultQueue, tuiQueue)
	go srv.Serve(lis)

	<-ctx.Done()
	srv.Stop()
}

// serverJob is what the server tracks about a job.
type serverJob struct {
	startTime time.Time     // startTime is the stamp this job was scheduled in the jobQueue.
	ttl       time.Duration // ttl is the allocated time for this job to finish.
	shim.Job
}

func (j serverJob) deadline() time.Time { return j.startTime.Add(j.ttl) }

type serverChunk struct {
	tuning.Range
	checksum  checksum.Checksum
	completed bool
	jobs      []serverJob
}

// deadline is the latest deadline of the deadlines of the chunks job. does
// not function properly if there are no jobs for the chunk, but we should
// schedule jobs on jobless chunks first anyway.
func (c serverChunk) deadline() time.Time {
	maxTime := time.Now().Add(time.Duration(-10) * time.Hour)
	for _, j := range c.jobs {
		if j.deadline().After(maxTime) {
			maxTime = j.deadline()
		}
	}
	return maxTime
}

type batchTracker []serverChunk

// completed determines if all chunks are completed in the batch.
func (b batchTracker) completed() bool {
	for _, chunk := range b {
		if !chunk.completed {
			return false
		}
	}
	return true
}

type match struct {
	chunkIx, jobIx int
}

func (b batchTracker) match(r shim.Result) (matched match, ok bool) {
	for i, chunk := range b {
		for j, job := range chunk.jobs {

			if job.UUID == r.UUID {
				return match{chunkIx: i, jobIx: j}, true
			}
		}
	}
	return match{}, false
}

// schedule contains our job scheduling rules.
//
//   - while there are jobless chunks we should immediately schedule those.
//   - if there is no job less chunk, we should only look at chunks that are not completed.
//   - if there is a non-completed chunk with jobs, we should find the chunk with earliest deadline.
//   - if the earliest deadline chunk deadline has passed schedule that one.
//   - otherwise no job to schedule.
func (b batchTracker) schedule() (chunkIx int, ok bool) {
	for i, chunk := range b {
		if len(chunk.jobs) == 0 {
			return i, true
		}
	}

	// there are no jobless chunks. find the earliest deadline
	minTime := time.Now().Add(time.Duration(10) * time.Hour)
	ix := -1
	for i, chunk := range b {
		if deadline := chunk.deadline(); !chunk.completed && chunk.deadline().Before(minTime) {
			minTime = deadline
			ix = i
		}
	}

	if ix != -1 && minTime.Before(time.Now()) {
		return ix, true
	}

	return -1, false
}

func epdProcess(
	ctx context.Context,
	fn,
	outFn string,
	k float64,
	jobQueue chan<- shim.Job,
	resultQueue <-chan shim.Result,
	tuiQueue chan<- tui.Update,
) {
	eCoeffs := tuning.EngineCoeffs()

	mse, err := fileMSE(ctx, fn, k, &eCoeffs)
	if err != nil {
		slog.Error("mse calculation error", "error", err)
		os.Exit(app.ExitFailure)
	}

	coeffs := eCoeffs.ToVector(tuning.DefaultTargets)
	momentum := tuning.NullVector(tuning.DefaultTargets)
	velocity := tuning.NullVector(tuning.DefaultTargets)
	lr := tuning.InitialLearningRate
	sumJobTimes := 0 * time.Second
	completeJobCnt := 0

	tuiQueue <- tui.LRUpdate{LR: lr}

	chunker, err := epd.NewChunker(fn)
	if err != nil {
		slog.Error("open error", "error", err)
		os.Exit(app.ExitFailure)
	}

	for epoch := 1; true; {
		tuiQueue <- tui.EpochUpdate{Epoch: epoch}

		for batch := range tuning.Batches(chunker.LineCount()) {
			tuiQueue <- tui.BatchUpdate{Start: batch.Start, End: batch.End}
			batchStart := time.Now()

			grads := tuning.NullVector(tuning.DefaultTargets)

			// gather the chunks in the batch and create server tracking structures
			tracker := make(batchTracker, 0, tuning.NumChunksInBatch)
			for chunk := range tuning.Chunks(batch) {
				tracker = append(tracker, serverChunk{Range: chunk})
			}

			// while there is an incomplete chunk in the batch
			for !tracker.completed() {
				if chunkIx, ok := tracker.schedule(); ok {
					chunk := &tracker[chunkIx]

					//create a sJob for the batch
					var ttl time.Duration
					if completeJobCnt == 0 {
						ttl = JobTTLInit * time.Second
					} else {
						// 2x the running average of our job times. If everything is in
						// order this should prevent us from scheduling an other job for
						// this chunk, but if something bad happens, we will schedule
						// another job once the deadline passes.
						ttl = 2 * sumJobTimes / time.Duration(completeJobCnt)
					}

					sJob := serverJob{
						startTime: time.Now(),
						ttl:       ttl,
						Job: shim.Job{
							UUID:         uuid.New(),
							Epoch:        epoch,
							Range:        chunk.Range,
							Coefficients: coeffs,
							K:            k,
						},
					}

					// put the job in the tracking structures
					chunk.jobs = append(chunk.jobs, sJob)

					// send the job to the tui
					tuiQueue <- tui.JobUpdate{
						ChunkIx:   chunkIx,
						JobIx:     len(chunk.jobs) - 1,
						StartTime: sJob.startTime,
						TTL:       sJob.ttl,
					}

					tuiQueue <- tui.MsgUpdate{Msg: "queueing job", Args: []any{"uuid", sJob.UUID}}

					// send the job to the client handler
					jobQueue <- sJob.Job
				}

				// register results
				select {
				case result := <-resultQueue:
					// validate result coming from client and search for a matching job in our structures
					tuiQueue <- tui.MsgUpdate{Msg: "received results", Args: []any{"uuid", result.UUID}}

					if match, ok := tracker.match(result); ok {
						chunk := &tracker[match.chunkIx]
						job := &chunk.jobs[match.jobIx]
						// if already completed ignore
						if !chunk.completed {
							grads.Add(result.Gradients)
							chunk.completed = true
							tuiQueue <- tui.ResultUpdate{ChunkIx: match.chunkIx, JobIx: match.jobIx}
						}
						// either way it's a data point for the running average of job times.
						sumJobTimes += time.Since(job.startTime)
						completeJobCnt++
					}

				case <-time.After(ClientWaitTime * time.Millisecond):
					// no results coming in yet

				case <-ctx.Done():
					return
				}
			}

			// batch completed, apply ADAM algo
			grads.DivConst(float64(batch.Len())) // average over the batch

			momentum.Combine(grads, func(m, g float64) float64 { return tuning.Beta1*m + (1-tuning.Beta1)*g })
			velocity.Combine(grads, func(v, g float64) float64 { return tuning.Beta2*v + (1-tuning.Beta2)*g*g })

			mHat := momentum.Map(func(m float64) float64 { return m / (1 - math.Pow(tuning.Beta1, float64(epoch))) })
			vHat := velocity.Map(func(v float64) float64 { return v / (1 - math.Pow(tuning.Beta2, float64(epoch))) })

			step := mHat
			step.Combine(vHat, func(mh, vh float64) float64 { return lr * mh / (1e-8 + math.Sqrt(vh)) })

			coeffs.Sub(step)

			tuiQueue <- tui.BatchTimeUpdate{Duration: time.Since(batchStart)}
		}

		// epoch completed, output coeffs and drop learning rate based on MSE change
		epoch++

		eCoeffs.SetVector(coeffs, tuning.DefaultTargets)
		newMSE, err := fileMSE(ctx, fn, k, &eCoeffs)
		if err != nil {
			slog.Error("mse calculation error", "error", err)
			os.Exit(app.ExitFailure)
		}

		f, err := os.Create(outFn)
		if err != nil {
			slog.Error("coeffs.go", "error", err)
			os.Exit(app.ExitFailure)
		}
		eCoeffs.Save(f, fn, epoch, newMSE)
		f.Close()

		tuiQueue <- tui.MSEUpdate{MSE: newMSE}

		if newMSE > mse {
			lr /= 2
			tuiQueue <- tui.LRUpdate{LR: lr}
		}
		mse = newMSE
	}
}

func minimizeK(ctx context.Context, fn string, tuiQueue chan<- tui.Update, pprofFile, memprofFile string) (float64, error) {
	if pprofFile != "" {
		f, err := os.Create(pprofFile)
		if err != nil {
			slog.Error("pprof file creation error", "error", err)
			os.Exit(app.ExitFailure)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			slog.Error("pprof start error", "error", err)
			os.Exit(app.ExitFailure)
		}
		defer pprof.StopCPUProfile()
	}

	tuiQueue <- tui.MsgUpdate{Msg: "minimising mse with k"}

	coeffs := tuning.EngineCoeffs()

	k := 2.832 // a scaling constant
	improved := true
	step := 1.0

	mse, err := fileMSE(ctx, fn, k, &coeffs)
	if err != nil {
		return 0, err
	}
	tuiQueue <- tui.KUpdate{K: k, Step: step}

	b := board.Board{}
	res := 0.0
	byLines, err := epd.OpenByLines(fn)
	if err != nil {
		return 0, err
	}
	defer byLines.Close()

	for step > 0.0001 {
		tuiQueue <- tui.KUpdate{K: k, Step: step}
		for improved {
			eHigh := 0.0
			eLow := 0.0
			cnt := 0
			for {
				line, err := byLines.Read()
				if err != nil {
					if err == io.EOF {
						byLines.Rewind()
						break
					}
					return 0, err
				}

				if err := epd.Parse(line, &b, &res); err != nil {
					return 0, err
				}
				score := coeffs.Eval(&b)
				sgm := tuning.Sigmoid(score, k+step)
				eHigh += (res - sgm) * (res - sgm)
				sgm = tuning.Sigmoid(score, k-step)
				eLow += (res - sgm) * (res - sgm)
				cnt++

				select {
				case <-ctx.Done():
					return 0, ctx.Err()

				default:
				}
			}
			if cnt == 0 {
				return 0, ErrInvalidEpd
			}
			eHigh /= float64(cnt)
			eLow /= float64(cnt)
			improved = false

			nK := k - step
			nE := eLow
			if eHigh < eLow {
				nK = k + step
				nE = eHigh
			}

			if nE < mse {
				improved = true
				mse = nE
				k = nK
				tuiQueue <- tui.MSEUpdate{MSE: mse}
				tuiQueue <- tui.KUpdate{K: k, Step: step}
			}
		}
		step /= 10.0
		improved = true
	}
	if memprofFile != "" {
		f, err := os.Create(memprofFile)
		if err != nil {
			slog.Error("memory profile error", "error", err)
			os.Exit(app.ExitFailure)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		// Lookup("allocs") creates a profile similar to go test -memprofile.
		// Alternatively, use Lookup("heap") for a profile
		// that has inuse_space as the default index.
		if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
			slog.Error("memory profile error", "error", err)
			os.Exit(app.ExitFailure)
		}
	}
	return k, nil
}

var ErrInvalidEpd = errors.New("invalid epd file")

func fileMSE(ctx context.Context, fn string, k float64, coeffs *tuning.EngineRep) (float64, error) {
	byLines, err := epd.OpenByLines(fn)
	if err != nil {
		return 0, err
	}
	defer byLines.Close()

	sum := 0.0
	cnt := 0
	b := board.Board{}
	res := 0.0

	for {
		line, err := byLines.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}

		err = epd.Parse(line, &b, &res)
		if err != nil {
			return 0, err
		}
		score := coeffs.Eval(&b)
		sgm := tuning.Sigmoid(score, k)
		sum += (res - sgm) * (res - sgm)
		cnt++

		select {
		case <-ctx.Done():
			return 0, ctx.Err()

		default:
		}
	}
	if cnt == 0 {
		slog.Error("no lines in epd", "filename", fn)
		return 0, ErrInvalidEpd
	}
	return sum / float64(cnt), nil
}
