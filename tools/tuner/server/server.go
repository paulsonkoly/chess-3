package server

import (
	"flag"
	"fmt"
	"iter"
	"log/slog"
	"net"
	"os"
	"time"

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
	tuning.Chunk
	completed bool
	jobs      []ServerJob
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

func (bc batchChunks) Match(r Result) (ix int, ok bool) {
	for i, chunk := range bc {
		for _, job := range chunk.jobs {
			if job.UUID == r.UUID {
				return i, true
			}
		}
	}
	return 0, false
}

type Job struct {
	UUID string
	tuning.Chunk
}

type ServerJob struct {
	deadline time.Time
	Job
}

type Result struct {
	UUID      string
	Gradients []float64
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

	jobQueue := make(chan Job, JobQueueDepth)
	resultQueue := make(chan Result, ResultQueueDepth)

	epd, err := epd.Load(epdFileName)
	if err != nil {
		slog.Error("failed to load epd file", "filename", epdFileName)
	}
	slog.Debug("loaded epd", "filename", epd.Basename(), "checksum", epd.Checksum)

	go epdProcess(epd, jobQueue, resultQueue)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		slog.Error("failed to bind port", "host", host, "port", port)
		os.Exit(1)
	}

	slog.Info("listening for incoming connections", "host", host, "port", port)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	s := tunerServer{
		jobQueue:    jobQueue,
		resultQueue: resultQueue,
		epdFilename: epd.Basename(),
		epdChecksum: epd.Checksum,
	}
	pb.RegisterTunerServer(grpcServer, s)
	grpcServer.Serve(lis)
}

func epdProcess(epd *epd.EPD, jobQueue chan<- Job, resultQueue <-chan Result) {
	coeffs := tuning.Coeffs{}

	for epoch := 1; true; epoch++ {
		slog.Debug("new epoch", "epoch", epoch)

		for batch := range epd.Batches() {

			grads := tuning.Coeffs{}

			// gather the chunks in the batch and create server tracking structures
			chunks := make(batchChunks, 0, tuning.NumChunksInBatch)
			for chunk := range batch.Chunks() {
				chunks = append(chunks, ServerChunk{Chunk: chunk})
			}

			// while there is an incomplete chunk in the batch
			for i, chunk := range chunks.Incomplete() {

				//create a job for the batch
				job := ServerJob{
					deadline: time.Now().Add(600 * time.Second),
					Job:      Job{Chunk: chunk.Chunk},
				}

				// put the job in the tracking structures
				chunks[i].jobs = append(chunks[i].jobs, job)

				slog.Debug("queueing job",
					"uuid", job.UUID,
					"deadline", job.deadline,
					"chunk.start", chunk.Start,
					"chunk.end", chunk.End)

				// send the job to the client handler
				jobQueue <- job.Job

				// register results
				select {
				case result := <-resultQueue:
					// validate result coming from client and search for a matching job in our structures
					if ix, ok := chunks.Match(result); ok {
						// if already completed ignore
						if !chunks[ix].completed {
							slog.Debug("received results", "uuid", result.UUID)
							grads.Add(result.Gradients)
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
		epd.Shuffle()
	}
}
