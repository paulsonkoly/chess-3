package server

import (
	"flag"
	"fmt"
	"log/slog"
	"math"
	"net"
	"os"
	"time"

	"github.com/essentialkaos/ek/v13/fmtutil/table"
	"github.com/google/uuid"
	"github.com/paulsonkoly/chess-3/tools/tuner/app"
	"github.com/paulsonkoly/chess-3/tools/tuner/checksum"
	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	"github.com/paulsonkoly/chess-3/tools/tuner/shim"
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

	sFlags := flag.NewFlagSet("server", flag.ExitOnError)
	sFlags.StringVar(&epdFileName, "epd", "", "epd file name")
	sFlags.StringVar(&host, "host", "localhost", "host to listen on")
	sFlags.IntVar(&port, "port", 9001, "port to listen on")
	sFlags.StringVar(&outFn, "out", "coeffs.go", "coeff output file")
	sFlags.Parse(args)

	eCoeffs := tuning.EngineCoeffs()
	eCoeffs.Save(os.Stdout, 1, 2.345)

	epdF, err := epd.New(epdFileName)
	if err != nil {
		slog.Error("failed to load epd file", "filename", epdFileName)
		os.Exit(app.ExitFailure)
	}
	slog.Debug("loaded epd", "filename", epdF.Basename())
	k, err := minimizeK(epdF)
	if err != nil {
		slog.Error("k minimization error", "error", err)
	}

	jobQueue := make(chan shim.Job, JobQueueDepth)
	resultQueue := make(chan shim.Result, ResultQueueDepth)

	go epdProcess(epdF, outFn, k, jobQueue, resultQueue)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		slog.Error("failed to bind port", "host", host, "port", port)
		os.Exit(app.ExitFailure)
	}

	slog.Info("listening for incoming connections", "host", host, "port", port)
	shim.NewServer(epdF, jobQueue, resultQueue).Serve(lis)
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
	chunk *serverChunk
	job   *serverJob
}

func (b batchTracker) match(r shim.Result) (matched *match, ok bool) {
	for i, chunk := range b {
		for j, job := range chunk.jobs {

			if job.UUID == r.UUID {
				return &match{chunk: &b[i], job: &b[i].jobs[j]}, true
			}
		}
	}
	return nil, false
}

// schedule contains our job scheduling rules.
//
//   - while there are jobless chunks we should immediately schedule those.
//   - if there is no job less chunk, we should only look at chunks that are not completed.
//   - if there is a non-completed chunk with jobs, we should find the chunk with earliest deadline.
//   - if the earliest deadline chunk deadline has passed schedule that one.
//   - otherwise no job to schedule.
func (b batchTracker) schedule() (chunk *serverChunk, ok bool) {
	for i, chunk := range b {
		if len(chunk.jobs) == 0 {
			return &b[i], true
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
		return &b[ix], true
	}

	return nil, false
}

func epdProcess(epdF *epd.File, outFn string, k float64, jobQueue chan<- shim.Job, resultQueue <-chan shim.Result) {
	eCoeffs := tuning.EngineCoeffs()

	mse, err := fileMSE(epdF, k, &eCoeffs)
	if err != nil {
		slog.Error("mse calculation error", "error", err)
	}

	coeffs := eCoeffs.ToVector(tuning.DefaultTargets)
	momentum := tuning.NullVector(tuning.DefaultTargets)
	velocity := tuning.NullVector(tuning.DefaultTargets)
	lr := tuning.InitialLearningRate
	sumJobTimes := 0 * time.Second
	completeJobCnt := 0

	for epoch := 1; true; {
		slog.Debug("new epoch", "epoch", epoch)

		for batch := range tuning.Batches(epdF.LineCount()) {

			grads := tuning.NullVector(tuning.DefaultTargets)

			// gather the chunks in the batch and create server tracking structures
			tracker := make(batchTracker, 0, tuning.NumChunksInBatch)
			for chunk := range tuning.Chunks(batch) {
				checksum, err := epdF.ChunkChecksum(epoch, chunk.Start, chunk.End)
				if err != nil {
					slog.Error("checksum calculation error", "error", err)
				}
				tracker = append(tracker, serverChunk{Range: chunk, checksum: checksum})
			}

			// while there is an incomplete chunk in the batch
			for !tracker.completed() {

				// TODO this is throw away code for now, I envisage some cool tui
				// interface for this
				tbl := table.NewTable()
				headers := make([]string, 0)
				maxJobs := 0
				for _, chunk := range tracker {
					if chunk.completed {
						headers = append(headers, "D") // done
					} else if len(chunk.jobs) == 0 {
						headers = append(headers, "N") // new
					} else {
						headers = append(headers, "P") // in progress
					}
					if len(chunk.jobs) > maxJobs {
						maxJobs = len(chunk.jobs)
					}
				}
				tbl.SetHeaders(headers...)

				lines := make([][]any, maxJobs)
				for i := range lines {
					lines[i] = make([]any, len(tracker))
				}
				now := time.Now()
				for cIx, chunk := range tracker {
					for jIx := range maxJobs {
						if jIx >= len(chunk.jobs) {
							lines[jIx][cIx] = "-"
						} else {
							diff := chunk.jobs[jIx].deadline().Sub(now)
							lines[jIx][cIx] = fmt.Sprintf("%.0f", diff.Seconds())
						}
					}
				}
				for _, line := range lines {
					tbl.Add(line...)
				}
				tbl.Render()

				if chunk, ok := tracker.schedule(); ok {
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
							Checksum:     chunk.checksum,
							Coefficients: coeffs,
							K:            k,
						},
					}

					// put the job in the tracking structures
					chunk.jobs = append(chunk.jobs, sJob)

					slog.Debug("queueing job", "job", sJob)

					// send the job to the client handler
					jobQueue <- sJob.Job
				}

				// register results
				select {
				case result := <-resultQueue:
					// validate result coming from client and search for a matching job in our structures
					slog.Debug("received results", "result", result)
					if match, ok := tracker.match(result); ok {
						// if already completed ignore
						if !match.chunk.completed {
							grads.Add(result.Gradients)
							match.chunk.completed = true
						}
						// either way it's a data point for the running average of job times.
						sumJobTimes += time.Since(match.job.startTime)
						completeJobCnt++
					}

				case <-time.After(ClientWaitTime * time.Millisecond):
					// no results coming in yet
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
		}

		// epoch completed, output coeffs and drop learning rate based on MSE change
		epoch++
		fmt.Println(epoch)

		eCoeffs.SetVector(coeffs, tuning.DefaultTargets)
		newMSE, err := fileMSE(epdF, k, &eCoeffs)
		if err != nil {
			slog.Error("mse calculation error", "error", err)
			os.Exit(app.ExitFailure)
		}

		f, err := os.Create(outFn)
		if err != nil {
			slog.Error("coeffs.go", "error", err)
			os.Exit(app.ExitFailure)
		}
		eCoeffs.Save(f, epoch, newMSE)
		f.Close()

		fmt.Printf("error drop %.10f , bestE %.10f\n", mse-newMSE, newMSE)
		if newMSE > mse {
			fmt.Printf("drop negative, LR %.4f -> %.4f\n", lr, lr/2.0)
			lr /= 2
		}
		mse = newMSE
	}
}

func minimizeK(epdF *epd.File) (float64, error) {
	coeffs := tuning.EngineCoeffs()

	k := 2.832 // a scaling constant
	improved := true
	step := 1.0

	mse, err := fileMSE(epdF, k, &coeffs)
	slog.Info("minimizing mse with k", "k", k, "mse", mse)
	if err != nil {
		return 0, err
	}
	for step > 0.0001 {
		slog.Info("k step", "k", k, "step", step)
		for improved {
			eHigh, err := fileMSE(epdF, k+step, &coeffs)
			if err != nil {
				return 0, err
			}
			eLow, err := fileMSE(epdF, k-step, &coeffs)
			if err != nil {
				return 0, err
			}
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
				slog.Info("minimizing mse with k", "k", k, "mse", mse)
			}
		}
		step /= 10.0
		improved = true
	}
	return k, nil
}

func fileMSE(epdF *epd.File, k float64, coeffs *tuning.EngineRep) (float64, error) {
	sum := float64(0)
	for batch := range tuning.Batches(epdF.LineCount()) {
		for chunk := range tuning.Chunks(batch) {
			chunkEntries, err := epdF.Chunk(1, chunk.Start, chunk.End)
			if err != nil {
				return 0, err
			}

			for _, epdE := range chunkEntries {
				b := epdE.Board
				r := epdE.Result

				score := coeffs.Eval(b)

				sgm := tuning.Sigmoid(score, k)

				sum += (r - sgm) * (r - sgm)
			}
		}
	}
	return sum / float64(epdF.LineCount()), nil
}
