package server

import (
	"flag"
	"fmt"
	"iter"
	"log/slog"
	"math"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/paulsonkoly/chess-3/tools/tuner/checksum"
	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	"github.com/paulsonkoly/chess-3/tools/tuner/shim"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
)

const (
	// JobTTL is job time to live duration in seconds.
	JobTTL = 600
	// JobQueueDepth determines how many jobs the EPD processor can produce
	// before its stalled back. Should not be smaller than NumChunksInBatch
	// otherwise we can create jobs for a whole batch simultaniously. It can be
	// bigger slightly, allowing jobs with passed TTL to live together with newly
	// dispatched jobs, in case a passed TTL job eventually finishes.
	JobQueueDepth    = 20
	ResultQueueDepth = 10
	// ClientWaitTime is the duration in ms the server waits for results before
	// creating new jobs.
	ClientWaitTime = 100
)

type ServerChunk struct {
	tuning.Range
	checksum  checksum.Checksum
	completed bool
	jobs      []serverJob
}

type batchChunks []ServerChunk

func (bc batchChunks) ConsistencyCheck() {
	// consistency check
	for _, c := range bc {
		if c.completed && len(c.jobs) == 0 {
			panic("chunk is marked completed, with no associated jobs")
		}
	}
}

func (bc batchChunks) Incomplete() iter.Seq2[int, ServerChunk] {
	return func(yield func(int, ServerChunk) bool) {

		// first look for a chunk that has no jobs
		for i, chunk := range bc {
			if len(chunk.jobs) == 0 {
				if !yield(i, chunk) {
					return
				}
			}
		}

		// all chunks must have jobs at this point
		allCompleted := false
		for !allCompleted {

			allCompleted = true

			// look for the chunk that has earliest deadline, if a chunk has multiple
			// jobs its deadline is the latest of its jobs deadlines. Ignore
			// completed chunks.
			minTime := time.Now().Add(time.Duration(10) * time.Hour)
			minIx := -1

			for i, chunk := range bc {
				if !chunk.completed {
					maxTime := time.Now().Add(time.Duration(-10) * time.Hour)
					for _, j := range chunk.jobs {
						if j.deadline.After(maxTime) {
							maxTime = j.deadline
						}
					}

					// TODO we can add extra conditions here that prevent scheduling jobs
					// for recently scheduled chunks.
					if maxTime.Before(minTime) {
						minTime = maxTime
						minIx = i
					}

					allCompleted = false
				}
			}

			if minIx != -1 {
				if !yield(minIx, bc[minIx]) {
					return
				}
			}
		}
	}
}

func (bc batchChunks) Match(r shim.Result) (ix int, ok bool) {
	for i, chunk := range bc {
		for _, job := range chunk.jobs {

			if job.UUID == r.UUID {
				return i, true
			}
		}
	}
	return 0, false
}

// serverJob is what the server tracks about a job
type serverJob struct {
	deadline time.Time
	shim.Job
}

func Run(args []string) {
	var epdFileName string
	var host string
	var port int

	sFlags := flag.NewFlagSet("server", flag.ExitOnError)
	sFlags.StringVar(&epdFileName, "epd", "", "epd file name")
	sFlags.StringVar(&host, "host", "localhost", "host to listen on")
	sFlags.IntVar(&port, "port", 9001, "port to listen on")
	sFlags.Parse(args)

	epdF, err := epd.Open(epdFileName)
	if err != nil {
		slog.Error("failed to load epd file", "filename", epdFileName)
		os.Exit(tuning.ExitFailure)
	}
	defer epdF.Close()
	checksum, err := epdF.Checksum()
	if err != nil {
		slog.Error("checksum calculation error", "error", err)
	}
	slog.Debug("loaded epd", "filename", epdF.Basename(), "checksum", checksum)
	k, err := minimizeK(epdF)
	if err != nil {
		slog.Error("k minimization error", "error", err)
	}

	jobQueue := make(chan shim.Job, JobQueueDepth)
	resultQueue := make(chan shim.Result, ResultQueueDepth)

	go epdProcess(epdF, k, jobQueue, resultQueue)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		slog.Error("failed to bind port", "host", host, "port", port)
		os.Exit(tuning.ExitFailure)
	}

	slog.Info("listening for incoming connections", "host", host, "port", port)
	shim.NewServer(epdF, jobQueue, resultQueue).Serve(lis)
}

func epdProcess(epdF *epd.File, k float64, jobQueue chan<- shim.Job, resultQueue <-chan shim.Result) {
	eCoeffs := tuning.EngineCoeffs()

	mse, err := fileMSE(epdF, k, &eCoeffs)
	if err != nil {
		slog.Error("mse calculation error", "error", err)
	}

	coeffs := eCoeffs.ToVector(tuning.DefaultTargets)
	momentum := tuning.NullVector(tuning.DefaultTargets)
	velocity := tuning.NullVector(tuning.DefaultTargets)
	lr := tuning.InitialLearningRate

	for epoch := 1; true; {
		slog.Debug("new epoch", "epoch", epoch)

		for batch := range tuning.Batches(epdF.LineCount()) {

			grads := tuning.NullVector(tuning.DefaultTargets)

			// gather the chunks in the batch and create server tracking structures
			chunks := make(batchChunks, 0, tuning.NumChunksInBatch)
			for chunk := range tuning.Chunks(batch) {
				checksum, err := epdF.ChunkChecksum(epoch , chunk.Start, chunk.End)
				if err != nil {
					slog.Error("checksum calculation error", "error", err)
				}
				chunks = append(chunks, ServerChunk{Range: chunk, checksum: checksum})
			}

			// while there is an incomplete chunk in the batch
			for i, chunk := range chunks.Incomplete() {

				//create a sJob for the batch
				sJob := serverJob{
					deadline: time.Now().Add(600 * time.Second),
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
				chunks[i].jobs = append(chunks[i].jobs, sJob)

				slog.Debug("queueing job", "job", sJob)

				// send the job to the client handler
				jobQueue <- sJob.Job

				// register results
				select {
				case result := <-resultQueue:
					// validate result coming from client and search for a matching job in our structures
					slog.Debug("received results", "result", result)
					if ix, ok := chunks.Match(result); ok {
						// if already completed ignore
						if !chunks[ix].completed {
							grads.Add(result.Gradients)
							chunks[ix].completed = true
						}
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
		fmt.Println(eCoeffs)

		newMSE, err := fileMSE(epdF, k, &eCoeffs)
		if err != nil {
			slog.Error("mse calculation error", "error", err)
			os.Exit(tuning.ExitFailure)
		}
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
