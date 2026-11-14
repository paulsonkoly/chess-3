package client

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/tools/tuner/app"
	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	"github.com/paulsonkoly/chess-3/tools/tuner/eval"
	"github.com/paulsonkoly/chess-3/tools/tuner/shim"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
	"golang.org/x/sync/errgroup"
)

var (
	errServerShutdown   = errors.New("server shutdown")
	errRetriesExhausted = errors.New("retries exhausted")
)

const (
	EPDRetryCount = 10
)

func Run(args []string) int {
	var host string
	var port int
	var numThreads int
	var workerPProf string
	var workerMProf string

	sFlags := flag.NewFlagSet("client", flag.ExitOnError)
	sFlags.StringVar(&host, "host", "localhost", "host to connect to")
	sFlags.IntVar(&port, "port", 9001, "port to connect to")
	sFlags.IntVar(&numThreads, "threads", runtime.NumCPU(), "number of worker threads")
	sFlags.StringVar(&workerPProf, "wpprof", "", "filename for gathering cpu profiling data for the first job of the worker")
	sFlags.StringVar(&workerMProf, "wmprof", "", "filename for gathering mem profiling data from the first job of the worker")
	sFlags.Parse(args)

	ctx := context.Background()

	client, err := shim.NewClient(host, port)
	if err != nil {
		slog.Error("error creating client", "error", err)
		return app.ExitFailure
	}
	defer client.Close()

	slog.Info("connected to server", "host", host, "port", port)

	epdInfo, err := client.RequestEPDInfo(ctx)
	if err != nil {
		slog.Error("failed requesting EPD info", "error", err)
		return app.ExitFailure
	}
	if err := obtainEPD(ctx, epdInfo, client); err != nil {
		slog.Error("failed obtaining EPD", "error", err)
		return app.ExitFailure
	}

	chunker, err := epd.NewChunker(epdInfo.Filename)
	if err != nil {
		slog.Error("chunker error", "error", err)
		return app.ExitFailure
	}

	eg, ctx := errgroup.WithContext(ctx)

	for ix := range numThreads {
		eg.Go(func() error {
			// only profile in the first worker
			if ix == 0 {
				return clientWorker(ctx, chunker, client, workerPProf, workerMProf)
			} else {
				return clientWorker(ctx, chunker, client, "", "")
			}
		})
	}

	if err := eg.Wait(); err != nil && !errors.Is(err, errServerShutdown) {
		slog.Error("client failed", "error", err)
		return app.ExitFailure
	}

	return 0
}

func obtainEPD(ctx context.Context, epdInfo shim.EPDInfo, client shim.Client) error {
	var lastErr error

	for retry := 0; retry < EPDRetryCount; retry++ {

		exists := false
		stat, err := os.Stat(epdInfo.Filename)
		if err == nil && stat.Mode().IsRegular() {
			exists = true
		}

		if !exists {
			slog.Info("downloading epd", "filename", epdInfo.Filename, "checksum", epdInfo.Checksum)
			if err := downloadEPD(ctx, epdInfo, client); err != nil {
				return err
			}
			continue
		}

		fChecksum, err := epd.Checksum(epdInfo.Filename)
		if err != nil {
			return fmt.Errorf("checksum %q: %w", epdInfo.Filename, err)
		}

		if epdInfo.Checksum.Matches(fChecksum) {
			return nil
		}

		lastErr = fmt.Errorf("checksum mismatch for %q", epdInfo.Filename)
		slog.Warn(
			"epd checksum mismatch",
			"filename", epdInfo.Filename,
			"epdF.Checksum", fChecksum,
			"epdInfo.Checksum", epdInfo.Checksum)

		slog.Debug("deleting local epd", "filename", epdInfo.Filename)
		if err := os.Remove(epdInfo.Filename); err != nil {
			return fmt.Errorf("remove bad epd %q: %w", epdInfo.Filename, err)
		}
	}

	if lastErr == nil {
		return errRetriesExhausted
	}
	return fmt.Errorf("%w after %d attempts: %w", errRetriesExhausted, EPDRetryCount, lastErr)
}

// downloadEPD streams the EPD file from the server. A failed write or read is an
// error: a partially written file would only be caught later by the checksum.
func downloadEPD(ctx context.Context, epdInfo shim.EPDInfo, client shim.Client) (err error) {
	stream, err := client.StreamEPD(ctx)
	if err != nil {
		return fmt.Errorf("stream epd: %w", err)
	}

	f, err := os.Create(epdInfo.Filename)
	if err != nil {
		return fmt.Errorf("create epd %q: %w", epdInfo.Filename, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("close epd %q: %w", epdInfo.Filename, cerr))
		}
	}()

	for {
		line, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("receive epd: %w", err)
		}

		if _, err := f.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("write epd: %w", err)
		}
	}
}

func clientWorker(ctx context.Context, chunker *epd.Chunker, client shim.Client, pprofFile, memprofFile string) error {
	ev := eval.New()
	for cnt := 0; ; cnt++ {
		slog.Debug("requesting job")
		job, shutdown, err := client.RequestJob(ctx)
		switch {
		case err != nil:
			if ctx.Err() != nil {
				// A sibling worker stopped or failed; it owns the real result.
				return nil
			}
			return fmt.Errorf("request job: %w", err)
		case shutdown:
			slog.Info("server shutdown")
			return errServerShutdown
		}
		slog.Info("received job", "job", job)

		var pprofOsFile *os.File

		if cnt == 0 {
			if pprofFile != "" {
				pprofOsFile, err = os.Create(pprofFile)
				if err != nil {
					return fmt.Errorf("create cpu profile %q: %w", pprofFile, err)
				}
				if err := pprof.StartCPUProfile(pprofOsFile); err != nil {
					return fmt.Errorf("start cpu profile: %w", err)
				}
			}
		}

		coeffs := job.Coefficients
		eCoeffs := tuning.EngineCoeffs()
		eCoeffs.SetVector(coeffs, tuning.DefaultTargets)
		grads := tuning.NullVector(tuning.DefaultTargets)
		k := job.K

		slog.Info("working on job", "job", job)

		b := board.Board{}
		res := 0.0

		fChunk, err := chunker.Open(job.Epoch, job.Range.Start, job.Range.End)
		if err != nil {
			return fmt.Errorf("open chunk epoch %d range %d-%d: %w", job.Epoch, job.Range.Start, job.Range.End, err)
		}
		chunkErr := func() error {
			defer fChunk.Close()

			for {
				if err := ctx.Err(); err != nil {
					return err
				}

				line, err := fChunk.Read()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					return fmt.Errorf("read chunk: %w", err)
				}
				if err := epd.Parse([]byte(line), &b, &res); err != nil {
					return fmt.Errorf("parse epd line %q: %w", line, err)
				}

				score := ev.Score(&b, (*eval.CoeffSet)(&eCoeffs))
				sigm := tuning.Sigmoid(score, k)
				loss := (res - sigm) * (res - sigm)

				for i, ptr := range eCoeffs.TunedParams(tuning.DefaultTargets) {
					old := *ptr
					*ptr += tuning.Epsilon

					score2 := ev.Score(&b, (*eval.CoeffSet)(&eCoeffs))
					sigm2 := tuning.Sigmoid(score2, k)
					loss2 := (res - sigm2) * (res - sigm2)

					grads.ModifyElem(i, func(g float64) float64 { return g + (loss2-loss)/tuning.Epsilon })
					*ptr = old
				}
			}
		}()
		if chunkErr != nil {
			if ctx.Err() != nil {
				// A sibling worker stopped or failed; it owns the real result.
				return nil
			}
			return chunkErr
		}

		if cnt == 0 {
			if memprofFile != "" {
				f, err := os.Create(memprofFile)
				if err != nil {
					return fmt.Errorf("create memory profile %q: %w", memprofFile, err)
				}
				runtime.GC() // get up-to-date statistics
				// Lookup("allocs") creates a profile similar to go test -memprofile.
				// Alternatively, use Lookup("heap") for a profile
				// that has inuse_space as the default index.
				if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
					return fmt.Errorf("write memory profile: %w", err)
				}
				if err := f.Close(); err != nil {
					return fmt.Errorf("close memory profile: %w", err)
				}
			}
			if pprofOsFile != nil {
				pprof.StopCPUProfile()
				if err := pprofOsFile.Close(); err != nil {
					return fmt.Errorf("close cpu profile: %w", err)
				}
			}
		}

		results := shim.Result{
			UUID:      job.UUID,
			Gradients: grads,
		}
		if err := client.RegisterResult(ctx, results); err != nil {
			if ctx.Err() != nil {
				// A sibling worker stopped or failed; it owns the real result.
				return nil
			}
			return fmt.Errorf("register result: %w", err)
		}
	}
}
