package client

import (
	"flag"
	"io"
	"log/slog"
	"os"
	"runtime"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/tools/tuner/app"
	"github.com/paulsonkoly/chess-3/tools/tuner/checksum"
	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	"github.com/paulsonkoly/chess-3/tools/tuner/shim"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
)

const (
	EPDRetryCount = 10
)

func Run(args []string) {
	var host string
	var port int
	var numThreads int

	sFlags := flag.NewFlagSet("client", flag.ExitOnError)
	sFlags.StringVar(&host, "host", "localhost", "host to connect to")
	sFlags.IntVar(&port, "port", 9001, "port to connect to")
	sFlags.IntVar(&numThreads, "threads", runtime.NumCPU(), "number of worker threads")
	sFlags.Parse(args)

	client, err := shim.NewClient(host, port)
	if err != nil {
		slog.Error("error creating client", "error", err)
		os.Exit(app.ExitFailure)
	}
	defer client.Close()

	slog.Info("connected to server", "host", host, "port", port)

	epdInfo := obtainEPDInfo(client)
	obtainEPD(epdInfo, client)

	chunker, err := epd.NewChunker(epdInfo.Filename)
	if err != nil {
		slog.Error("chunker error", "error", err)
		os.Exit(app.ExitFailure)
	}

	for range numThreads {
		go clientWorker(chunker, client)
	}

	select {}
}

func obtainEPDInfo(client shim.Client) shim.EPDInfo {
	slog.Debug("requesting epd info")
	epdInfo, err := client.RequestEPDInfo()
	if err != nil {
		slog.Error("failed requesting epd info", "error", err)
		os.Exit(app.ExitFailure)
	}
	return epdInfo
}

func obtainEPD(epdInfo shim.EPDInfo, client shim.Client) {
	var retry int

	for haveEPD := false; !haveEPD && retry < EPDRetryCount; {

		exists := false
		stat, err := os.Stat(epdInfo.Filename)
		if err == nil && stat.Mode().IsRegular() {
			exists = true
		}

		if !exists {
			slog.Info("downloading epd", "filename", epdInfo.Filename, "checksum", epdInfo.Checksum)

			stream, err := client.StreamEPD()
			if err != nil {
				slog.Error("stream error", "error", err)
				os.Exit(app.ExitFailure)
			}

			f, err := os.Create(epdInfo.Filename)
			if err != nil {
				slog.Error("file creation error", "error", err, "filename", epdInfo.Filename)
				os.Exit(app.ExitFailure)
			}

			for {
				line, err := stream.Recv()
				if err != nil {
					if err != io.EOF {
						slog.Warn("stream error", "error", err)
					}
					break
				}

				_, err = f.WriteString(line + "\n")
				if err != nil {
					slog.Warn("write error", "error", err)
				}
			}

			if err := f.Close(); err != nil {
				slog.Error("error closing file", "error", err)
				os.Exit(app.ExitFailure)
			}
		} else {
			fChecksum, err := epd.Checksum(epdInfo.Filename)
			if err != nil {
				slog.Error("checksum calculation error", "error", err)
				os.Exit(app.ExitFailure)
			}

			if !epdInfo.Checksum.Matches(fChecksum) {
				slog.Warn(
					"epd checksum mismatch",
					"filename", epdInfo.Filename,
					"epdF.Checksum", fChecksum,
					"epdInfo.Checksum", epdInfo.Checksum)

				slog.Debug("deleting local epd", "filename", epdInfo.Filename)
				if err := os.Remove(epdInfo.Filename); err != nil {
					slog.Error("can't remove bad epd file", "error", err)
					os.Exit(app.ExitFailure)
				}
			} else {
				return
			}
		}

		retry++
	}
	slog.Error("can't retrieve epd", "retry", retry)
	os.Exit(app.ExitFailure)
}

func clientWorker(chunker *epd.Chunker, client shim.Client) {
	for {
		slog.Debug("requesting job")
		job, err := client.RequestJob()
		if err != nil {
			slog.Error("job request error", "error", err)
			os.Exit(app.ExitFailure)
		}
		slog.Info("received job", "job", job)

		fChunk,err := chunker.Open(job.Epoch, job.Range.Start, job.Range.End)
		if err != nil {
			slog.Error("chunk mapping error", "error", err)
			os.Exit(app.ExitFailure)
		}

		cSumCol := checksum.NewCollector()
		for {
			line, err := fChunk.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				slog.Warn("read error", "error", err)
				continue
			}

			cSumCol.Collect(line)
		}

		if !job.Checksum.Matches(cSumCol.Checksum()) {
			slog.Error("chunk checksum mismatch", "checksum", cSumCol.Checksum(), "job.checksum", job.Checksum)
			os.Exit(app.ExitFailure)
		}

		coeffs := job.Coefficients
		eCoeffs := tuning.EngineCoeffs()
		eCoeffs.SetVector(coeffs, tuning.DefaultTargets)
		grads := tuning.NullVector(tuning.DefaultTargets)
		k := job.K

		slog.Info("working on job", "job", job)

		fChunk.Rewind()
		b := board.Board{}
		res := 0.0

		for {
			line, err := fChunk.Read()
			if err != nil {
				if err == io.EOF {
					fChunk.Close()
					break
				}
				slog.Warn("read error", "error", err)
				continue
			}
			if err := epd.Parse([]byte(line), &b, &res); err != nil {
				slog.Warn("parse error", "error", err)
				continue
			}

			score := eCoeffs.Eval(&b)
			sigm := tuning.Sigmoid(score, k)
			loss := (res - sigm) * (res - sigm)

			grads.CombinePerturbed(coeffs, tuning.Epsilon,
				func(g float64, c tuning.Vector) float64 {
					// we need to work on a local copy of eCoeffs.
					eCoeffs := eCoeffs
					eCoeffs.SetVector(c, tuning.DefaultTargets)

					score2 := eCoeffs.Eval(&b)
					sigm2 := tuning.Sigmoid(score2, k)
					loss2 := (res - sigm2) * (res - sigm2)

					return g + (loss2-loss)/tuning.Epsilon
				})
		}

		results := shim.Result{
			UUID:      job.UUID,
			Gradients: grads,
		}
		if err := client.RegisterResult(results); err != nil {
			slog.Warn("failed to register results", "error", err)
		}
	}
}
