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
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"slices"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/tools/tuner/app"
	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	"github.com/paulsonkoly/chess-3/tools/tuner/eval"
	"github.com/paulsonkoly/chess-3/tools/tuner/shim"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
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

func Run(args []string) int {
	var (
		epdFileName string
		host        string
		port        int
		outFn       string
		minKPProf   string
		minKMProf   string
		maxEpochs   int
		metricsPort int
	)

	sFlags := flag.NewFlagSet("server", flag.ExitOnError)
	sFlags.StringVar(&epdFileName, "epd", "", "epd file name")
	sFlags.StringVar(&host, "host", "localhost", "host to listen on")
	sFlags.IntVar(&port, "port", 9001, "port to listen on")
	sFlags.StringVar(&outFn, "out", "coeffs.go", "coeff output file")
	sFlags.StringVar(&minKPProf, "kpprof", "", "filename for gathering cpu profiling data from the mse minimization with k")
	sFlags.StringVar(&minKMProf, "kmprof", "", "filename for gathering mem profiling data from the mse minimization with k")
	sFlags.IntVar(&maxEpochs, "maxEpochs", 0, "number of epochs (0 disables)")
	sFlags.IntVar(&metricsPort, "metricsPort", 9111, "port for prometheus metrics endpoint (0 disables)")
	sFlags.Parse(args)

	if epdFileName == "" {
		slog.Error("no epd file given")
		return app.ExitFailure
	}

	switch {
	case maxEpochs < 0:
		slog.Error("maxEpochs negative", slog.Int("maxEpochs", maxEpochs))
		return app.ExitFailure
	case maxEpochs == 0:
		maxEpochs = math.MaxInt
	}

	if metricsPort < 0 {
		slog.Error("metricsPort negative", slog.Int("metricsPort", metricsPort))
		return app.ExitFailure
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx, sigCleanup := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer sigCleanup()

	eg := errgroup.Group{}

	if metricsPort != 0 {
		metricsLis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, metricsPort))
		if err != nil {
			slog.Error("failed to bind metrics port", "host", host, "port", metricsPort, "error", err)
			return app.ExitFailure
		}
		metricsSrv := &http.Server{
			Handler:           promhttp.Handler(),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       5 * time.Second,
		}

		eg.Go(func() error {
			if err := metricsSrv.Serve(metricsLis); err != nil && !errors.Is(err, http.ErrServerClosed) {
				cancel()
				return fmt.Errorf("metrics serve: %w", err)
			}
			return nil
		})

		eg.Go(func() error {
			<-ctx.Done()

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
				metricsSrv.Close()
				return fmt.Errorf("metrics shutdown: %w", err)
			}
			return nil
		})
	}

	k, err := minimizeK(ctx, epdFileName, minKPProf, minKMProf)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return 0
		}
		slog.Error("k minimization error", "error", err)
		return app.ExitFailure
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		slog.Error("failed to bind port", "host", host, "port", port, "error", err)
		return app.ExitFailure
	}
	slog.Info("listening for incoming connections", slog.String("host", host), slog.Int("port", port))
	tunerReady.Set(1)

	jobQueue := make(chan shim.Job, JobQueueDepth)
	resultQueue := make(chan shim.Result, ResultQueueDepth)
	shutdown := make(chan struct{})

	eg.Go(func() error {
		defer close(shutdown)
		err := epdProcess(ctx, epdFileName, outFn, k, maxEpochs, jobQueue, resultQueue)
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("epd process: %w", err)
		}
		return nil
	})

	srv := shim.NewServer(epdFileName, jobQueue, resultQueue, shutdown)
	eg.Go(func() error {
		defer cancel()
		defer tunerReady.Set(0)
		if err := srv.Serve(lis); err != nil {
			return fmt.Errorf("grpc serve: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		<-shutdown
		finished := make(chan struct{})
		go func() {
			srv.GracefulStop()
			close(finished)
		}()

		select {
		case <-time.After(5 * time.Second):
			srv.Stop()
		case <-finished:
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		slog.Error("server failed", "error", err)
		return app.ExitFailure
	}

	return 0
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
	maxEpochs int,
	jobQueue chan<- shim.Job,
	resultQueue <-chan shim.Result,
) error {
	eCoeffs := tuning.EngineCoeffs()

	mse, err := fileMSE(ctx, fn, k, &eCoeffs)
	if err != nil {
		return fmt.Errorf("mse calculation: %w", err)
	}

	coeffs := eCoeffs.ToVector(tuning.DefaultTargets)
	momentum := tuning.NullVector(tuning.DefaultTargets)
	velocity := tuning.NullVector(tuning.DefaultTargets)
	lr := tuning.InitialLearningRate
	sumJobTimes := 0 * time.Second
	completeJobCnt := 0

	mseBest := mse
	epochsSinceImprovement := 0
	stableEpochs := 0

	// The initial MSE is the top of the curve, recorded as epoch 0.
	tunerEpoch.Set(0)
	tunerMSE.Set(mse)
	tunerMSEInitial.Set(mse)
	tunerMSEBest.Set(mse)
	tunerMSEDelta.Set(0)
	tunerMSERelativeDelta.Set(0)
	tunerEpochsSinceImprovement.Set(0)
	tunerLearningRate.Set(lr)
	tunerK.Set(k)
	tunerCoeffsChanged.Set(0)
	tunerCoeffsChangedRatio.Set(0)
	tunerCoeffsStableEpochs.Set(0)
	tunerTunedCoeffs.Set(float64(len(coeffs.VectorToSlice())))

	slog.Info("LR update", slog.Float64("LR", lr))

	chunker, err := epd.NewChunker(fn)
	if err != nil {
		return fmt.Errorf("epd file open: %w", err)
	}

	numBatches := 0
	for range tuning.Batches(chunker.LineCount()) {
		numBatches++
	}
	tunerBatchesPerEpoch.Set(float64(numBatches))

	for epoch := 1; epoch <= maxEpochs; {
		slog.Info("epoch update", slog.Int("epoch", epoch))

		epochStart := time.Now()
		batchesDone := 0
		tunerEpochProgressRatio.Set(0)
		epochCoeffs := slices.Clone(coeffs.VectorToSlice())

		for batch := range tuning.Batches(chunker.LineCount()) {
			slog.Info("batch update", slog.Int("start", batch.Start), slog.Int("end", batch.End))
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

					slog.Debug("queueing job",
						slog.String("uuid", sJob.UUID.String()),
						slog.Int("chunkIx", chunkIx),
						slog.Int("jobIx", len(chunk.jobs)-1),
						slog.Time("startTime", sJob.startTime),
						slog.Duration("ttl", sJob.ttl),
					)

					select {
					case jobQueue <- sJob.Job:
						// send the job to the client handler
						tunerJobsScheduled.Inc()
					case <-ctx.Done():
						return ctx.Err()
					}
				}

				// register results
				select {
				case result := <-resultQueue:
					// validate result coming from client and search for a matching job in our structures
					slog.Debug("received results", slog.String("uuid", result.UUID.String()))

					if match, ok := tracker.match(result); ok {
						chunk := &tracker[match.chunkIx]
						job := &chunk.jobs[match.jobIx]
						// if already completed ignore
						if !chunk.completed {
							grads.Add(result.Gradients)
							chunk.completed = true
							tunerChunksCompleted.Inc()
							slog.Debug("result matched job", slog.Int("chunkIx", match.chunkIx), slog.Int("jobIx", match.jobIx))
						}
						// either way it's a data point for the running average of job times.
						sumJobTimes += time.Since(job.startTime)
						completeJobCnt++
						tunerSchedulerJobTimeAvg.Set((sumJobTimes / time.Duration(completeJobCnt)).Seconds())
					}

				case <-time.After(ClientWaitTime * time.Millisecond):
					// no results coming in yet

				case <-ctx.Done():
					return ctx.Err()
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

			batchDuration := time.Since(batchStart)
			tunerBatchDuration.Set(batchDuration.Seconds())
			slog.Info("batch time update", slog.Duration("duration", batchDuration))

			batchesDone++
			if numBatches > 0 {
				tunerEpochProgressRatio.Set(float64(batchesDone) / float64(numBatches))
			}
			tunerPositionsProcessed.Add(float64(batch.Len()))
		}

		// Count the net rounded change over the epoch. The generated
		// coefficients file stores integers, so compare what is written, not
		// the raw floating point values.
		changed := 0
		finalCoeffs := coeffs.VectorToSlice()
		for i, prev := range epochCoeffs {
			if tuning.RoundCoeff(prev) != tuning.RoundCoeff(finalCoeffs[i]) {
				changed++
			}
		}
		tuned := len(epochCoeffs)
		changedRatio := 0.0
		if tuned > 0 {
			changedRatio = float64(changed) / float64(tuned)
		}

		eCoeffs.SetVector(coeffs, tuning.DefaultTargets)
		newMSE, err := fileMSE(ctx, fn, k, &eCoeffs)
		if err != nil {
			return fmt.Errorf("mse calculation: %w", err)
		}

		if err := saveCoeffs(fn, outFn, epoch, newMSE, eCoeffs); err != nil {
			return fmt.Errorf("coeff saving: %w", err)
		}

		tunerEpoch.Set(float64(epoch))
		tunerMSE.Set(newMSE)
		tunerEpochsCompleted.Inc()
		tunerEpochDuration.Set(time.Since(epochStart).Seconds())

		// Coefficient churn is published only after the save succeeded, so the
		// chart always describes the coefficients file on disk.
		if changed == 0 {
			stableEpochs++
		} else {
			stableEpochs = 0
		}
		tunerCoeffsChanged.Set(float64(changed))
		tunerCoeffsChangedRatio.Set(changedRatio)
		tunerCoeffsStableEpochs.Set(float64(stableEpochs))

		delta := newMSE - mse
		tunerMSEDelta.Set(delta)

		relDelta := 0.0
		if mse != 0 {
			relDelta = delta / mse
		}
		tunerMSERelativeDelta.Set(relDelta)

		if newMSE < mseBest {
			mseBest = newMSE
			epochsSinceImprovement = 0
			tunerMSEBest.Set(mseBest)
		} else {
			epochsSinceImprovement++
		}
		tunerEpochsSinceImprovement.Set(float64(epochsSinceImprovement))

		// epoch completed, output coeffs and drop learning rate based on MSE change
		epoch++

		slog.Info("mse update", slog.Float64("mse", newMSE))

		if newMSE > mse {
			lr /= 2
			tunerLearningRate.Set(lr)
			slog.Info("lr update", slog.Float64("lr", lr))
		}
		mse = newMSE
	}
	return nil
}

func minimizeK(ctx context.Context, fn string, pprofFile, memprofFile string) (float64, error) {
	if pprofFile != "" {
		f, err := os.Create(pprofFile)
		if err != nil {
			return 0, fmt.Errorf("create cpu profile %q: %w", pprofFile, err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return 0, fmt.Errorf("start cpu profile: %w", err)
		}
		defer pprof.StopCPUProfile()
	}

	slog.Info("minimising mse with K")

	ev := eval.New()
	coeffs := tuning.EngineCoeffs()

	k := 2.832 // a scaling constant
	tunerK.Set(k)
	improved := true
	step := 1.0

	mse, err := fileMSE(ctx, fn, k, &coeffs)
	if err != nil {
		return 0, fmt.Errorf("mse calculation: %w", err)
	}
	slog.Info("K update", slog.Float64("K", k), slog.Float64("step", step))

	b := board.Board{}
	res := 0.0
	byLines, err := epd.OpenByLines(fn)
	if err != nil {
		return 0, fmt.Errorf("open epd %q: %w", fn, err)
	}
	defer byLines.Close()

	for step > 0.0001 {
		slog.Info("K update", slog.Float64("K", k), slog.Float64("step", step))
		for improved {
			eHigh := 0.0
			eLow := 0.0
			cnt := 0
			for {
				line, err := byLines.Read()
				if err != nil {
					if errors.Is(err, io.EOF) {
						if err := byLines.Rewind(); err != nil {
							return 0, fmt.Errorf("rewind epd %q: %w", fn, err)
						}
						break
					}
					return 0, fmt.Errorf("read epd %q: %w", fn, err)
				}

				if err := epd.Parse(line, &b, &res); err != nil {
					return 0, fmt.Errorf("parse epd line %q: %w", line, err)
				}
				score := ev.Score(&b, (*eval.CoeffSet)(&coeffs))
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
				return 0, fmt.Errorf("%w: %q", errInvalidEpd, fn)
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
				tunerK.Set(k)
				slog.Info("mse update", slog.Float64("mse", mse))
				slog.Info("K update", slog.Float64("K", k), slog.Float64("step", step))
			}
		}
		step /= 10.0
		improved = true
	}
	if memprofFile != "" {
		f, err := os.Create(memprofFile)
		if err != nil {
			return 0, fmt.Errorf("create memory profile %q: %w", memprofFile, err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		// Lookup("allocs") creates a profile similar to go test -memprofile.
		// Alternatively, use Lookup("heap") for a profile
		// that has inuse_space as the default index.
		if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
			return 0, fmt.Errorf("write memory profile: %w", err)
		}
	}
	return k, nil
}

var errInvalidEpd = errors.New("invalid epd file")

func fileMSE(ctx context.Context, fn string, k float64, coeffs *tuning.EngineRep) (float64, error) {
	byLines, err := epd.OpenByLines(fn)
	if err != nil {
		return 0, fmt.Errorf("open epd %q: %w", fn, err)
	}
	defer byLines.Close()

	sum := 0.0
	cnt := 0
	b := board.Board{}
	res := 0.0

	ev := eval.New()

	for {
		line, err := byLines.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, fmt.Errorf("read epd %q: %w", fn, err)
		}

		err = epd.Parse(line, &b, &res)
		if err != nil {
			return 0, fmt.Errorf("parse epd line %q: %w", line, err)
		}
		score := ev.Score(&b, (*eval.CoeffSet)(coeffs))
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
		return 0, fmt.Errorf("%w: %q", errInvalidEpd, fn)
	}
	return sum / float64(cnt), nil
}

func saveCoeffs(epdFn, outFn string, epoch int, mse float64, eCoeffs tuning.EngineRep) error {
	dir := filepath.Dir(outFn)
	tmp, err := os.CreateTemp(dir, "")
	if err != nil {
		return fmt.Errorf("create temp file in %q: %w", dir, err)
	}
	// this fails on the normal path as the file is already gone. This is here to
	// ensures that we never leave a temp file behind in case of error - unless
	// our process is killed.
	defer func() { _ = os.Remove(tmp.Name()) }()

	if err := eCoeffs.Save(tmp, epdFn, epoch, mse); err != nil {
		return fmt.Errorf("write coeffs: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync coeffs %q: %w", tmp.Name(), err)
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return fmt.Errorf("chmod coeffs %q: %w", tmp.Name(), err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close coeffs %q: %w", tmp.Name(), err)
	}
	if err := os.Rename(tmp.Name(), outFn); err != nil {
		return fmt.Errorf("rename coeffs to %q: %w", outFn, err)
	}

	return nil
}
