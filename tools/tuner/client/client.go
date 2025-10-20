package client

import (
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	"github.com/paulsonkoly/chess-3/tools/tuner/shim"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
	"golang.org/x/sys/unix"
)

func Run(args []string) {
	var host string
	var port int

	sFlags := flag.NewFlagSet("server", flag.ExitOnError)
	sFlags.StringVar(&host, "host", "localhost", "host to connect to")
	sFlags.IntVar(&port, "port", 9001, "port to connect to")
	sFlags.Parse(args)

	client, err := shim.NewClient(host, port)
	if err != nil {
		slog.Error("error creating client", "error", err)
	}
	defer client.Close()
	slog.Info("connected to server", "host", host, "port", port)

	slog.Debug("requesting epd info")
	epdInfo, err := client.RequestEPDInfo()
	if err != nil {
		slog.Error("failed requesting epd info", "error", err)
		os.Exit(tuning.ExitFailure)
	}

	var epdF *epd.File

	for haveEPD := false; !haveEPD; {
		epdF, err = epd.Open(epdInfo.Filename)
		if err != nil {
			if !errors.Is(err, unix.ENOENT) {
				slog.Error("unexpected error on epd load", "error", err)
				os.Exit(tuning.ExitFailure)
			}

			slog.Info("downloading epd", "filename", epdInfo.Filename, "checksum", epdInfo.Checksum)

			stream, err := client.StreamEPD()
			if err != nil {
				slog.Error("stream error", "error", err)
				os.Exit(tuning.ExitFailure)
			}

			f, err := os.Create(epdInfo.Filename)
			if err != nil {
				slog.Error("file creation error", "error", err, "filename", epdInfo.Filename)
				os.Exit(tuning.ExitFailure)
			}

			for {
				line, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					slog.Warn("stream error", "error", err)
				}

				_, err = f.WriteString(line + "\n")
				if err != nil {
					slog.Warn("write error", "error", err)
				}
			}

			f.Close()
		} else {
			fChecksum, err := epdF.Checksum()
			if err != nil {
				slog.Error("checksum calculation error", "error", err)
				os.Exit(tuning.ExitFailure)
			}

			if !epdInfo.Checksum.Matches(fChecksum) {
				epdF.Close()

				slog.Warn(
					"epd checksum mismatch",
					"filename", epdInfo.Filename,
					"epdF.Checksum", fChecksum,
					"epdInfo.Checksum", epdInfo.Checksum)

				slog.Debug("deleting local epd", "filename", epdF.Basename())
				if err := os.Remove(epdF.Basename()); err != nil {
					slog.Error("can't remove bad epd file", "error", err)
					os.Exit(tuning.ExitFailure)
				}
			} else {
				haveEPD = true
				defer epdF.Close()
			}
		}
	}

	epoch := 1

	for {
		slog.Debug("requesting job")
		job, err := client.RequestJob()
		if err != nil {
			slog.Error("job request error", "error", err)
			os.Exit(tuning.ExitFailure)
		}
		slog.Info("received job", "job", job)

		if int(job.Epoch) != epoch {
			slog.Info("shuffling epd", "epoch", job.Epoch)
			epoch = job.Epoch
			epdF.Shuffle(epoch)
		}

		checksum, err := epdF.ChunkChecksum(job.Range.Start, job.Range.End)
		if err != nil {
			slog.Error("checksum calculation error", "error", err)
			os.Exit(tuning.ExitFailure)
		}

		if !job.Checksum.Matches(checksum) {
			slog.Warn("chunk checksum mismatch") // TODO args
			os.Exit(tuning.ExitFailure)
		}

		chunk, err := epdF.Chunk(job.Range.Start, job.Range.End)
		if err != nil {
			slog.Error("chunking error", "error", err)
			os.Exit(tuning.ExitFailure)
		}

		coeffs := job.Coefficients
		eCoeffs := tuning.EngineCoeffs()
		eCoeffs.SetVector(coeffs, tuning.DefaultTargets)
		grads := tuning.NullVector(tuning.DefaultTargets)
		k := job.K

		slog.Info("working on job", "job", job)

		for _, entry := range chunk {
			score := eCoeffs.Eval(entry.Board)
			sigm := tuning.Sigmoid(score, k)
			loss := (entry.Result - sigm) * (entry.Result - sigm)

			grads.CombinePerturbed(coeffs, tuning.Epsilon,
				func(g float64, c tuning.Vector) float64 {
					eCoeffs.SetVector(c, tuning.DefaultTargets)

					score2 := eCoeffs.Eval(entry.Board)
					sigm2 := tuning.Sigmoid(score2, k)
					loss2 := (entry.Result - sigm2) * (entry.Result - sigm2)

					return g + (loss2-loss)/tuning.Epsilon
				})
		}

		// TODO
		result := shim.Result{
			UUID:      job.UUID,
			Gradients: grads,
		}

		slog.Info("sending result", "result", result)

		client.RegisterResult(result)
	}
}
