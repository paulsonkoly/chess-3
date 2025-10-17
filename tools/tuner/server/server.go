package server

import (
	"encoding/base64"
	"flag"
	"fmt"
	"iter"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
	"google.golang.org/grpc"
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
	checksum  []byte
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

func (bc batchChunks) Match(r result) (ix int, ok bool) {
	for i, chunk := range bc {
		for _, job := range chunk.jobs {

			if job.uuid == r.uuid {
				return i, true
			}
		}
	}
	return 0, false
}

// queueJob is what the server emits to the grpc shim for serialisation
type queueJob struct {
	uuid     uuid.UUID
	epoch    int
	start    int
	end      int
	checksum []byte
	k        float64
}

func (qj queueJob) String() string {
	return fmt.Sprintf("{uuid = %s, epoch = %d, start = %d, end = %d, checksum = %s, k = %f}",
		qj.uuid.String(),
		qj.epoch,
		qj.start,
		qj.end,
		base64.URLEncoding.EncodeToString(qj.checksum),
		qj.k)
}

// serverJob is what the server tracks about a job
type serverJob struct {
	deadline time.Time
	queueJob
}

// result is what we receive deserialised from the grpc shim
type result struct {
	uuid      uuid.UUID
	gradients []float64
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
	slog.Debug("loaded epd", "filename", epdF.Basename(), "checksum", base64.URLEncoding.EncodeToString(checksum))
	k, err := minimizeK(epdF)
	if err != nil {
		slog.Error("k minimization error", "error", err)
	}

	jobQueue := make(chan queueJob, JobQueueDepth)
	resultQueue := make(chan result, ResultQueueDepth)

	go epdProcess(epdF, k, jobQueue, resultQueue)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		slog.Error("failed to bind port", "host", host, "port", port)
		os.Exit(tuning.ExitFailure)
	}

	slog.Info("listening for incoming connections", "host", host, "port", port)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	s := tunerServer{
		jobQueue:    jobQueue,
		resultQueue: resultQueue,
		epdF:        epdF,
	}
	pb.RegisterTunerServer(grpcServer, s)
	grpcServer.Serve(lis)
}

func epdProcess(epdF *epd.File, k float64, jobQueue chan<- queueJob, resultQueue <-chan result) {
	coeffs, err := tuning.EngineCoeffs()
	if err != nil {
		slog.Error("error loading engine coeffs", "error", err)
		os.Exit(tuning.ExitFailure)
	}

	for epoch := 1; true; epoch++ {
		slog.Debug("new epoch", "epoch", epoch)

		for batch := range tuning.Batches(epdF.LineCount()) {

			grads := tuning.Coeffs{}

			// gather the chunks in the batch and create server tracking structures
			chunks := make(batchChunks, 0, tuning.NumChunksInBatch)
			for chunk := range tuning.Chunks(batch) {
				checksum, err := epdF.ChunkChecksum(chunk.Start, chunk.End)
				if err != nil {
					slog.Error("checksum calculation error", "error", err)
				}
				chunks = append(chunks, ServerChunk{Range: chunk, checksum: checksum})
			}

			// while there is an incomplete chunk in the batch
			for i, chunk := range chunks.Incomplete() {

				//create a job for the batch
				job := serverJob{
					deadline: time.Now().Add(600 * time.Second),
					queueJob: queueJob{
						uuid:     uuid.New(),
						epoch:    epoch,
						start:    chunk.Start,
						end:      chunk.End,
						checksum: chunk.checksum,
						k:        k,
					},
				}

				// put the job in the tracking structures
				chunks[i].jobs = append(chunks[i].jobs, job)

				slog.Debug("queueing job", "job", job.queueJob)

				// send the job to the client handler
				jobQueue <- job.queueJob

				// register results
				select {
				case result := <-resultQueue:
					// validate result coming from client and search for a matching job in our structures
					if ix, ok := chunks.Match(result); ok {
						// if already completed ignore
						if !chunks[ix].completed {
							slog.Debug("received results", "uuid", result.uuid)
							// grads.Add(result.gradients)
							chunks[ix].completed = true
						}
					}

				case <-time.After(ClientWaitTime * time.Millisecond):
					// no results coming in yet
				}
			}

			// batch completed, add the gradient vector to the coeffs
			coeffs.Add(grads)
		}

		// epoch completed, output coeffs
		fmt.Println(coeffs)

		// shuffle
		epdF.Shuffle(epoch)
	}
}

func minimizeK(epdF *epd.File) (float64, error) {
	coeffs, err := tuning.EngineCoeffs()
	if err != nil {
		return 0, err
	}

	k := 2.839 // a scaling constant
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

func fileMSE(epdF *epd.File, k float64, coeffs *tuning.Coeffs) (float64, error) {
	sum := float64(0)
	for batch := range tuning.Batches(epdF.LineCount()) {
		for chunk := range tuning.Chunks(batch) {
			chunkEntries, err := epdF.Chunk(chunk.Start, chunk.End)
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
