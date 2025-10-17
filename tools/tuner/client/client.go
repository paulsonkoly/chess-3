package client

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"

	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
	"github.com/paulsonkoly/chess-3/tools/tuner/tuning"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Run(args []string) {
	var host string
	var port int

	sFlags := flag.NewFlagSet("server", flag.ExitOnError)
	sFlags.StringVar(&host, "host", "localhost", "host to connect to")
	sFlags.IntVar(&port, "port", 9001, "port to connect to")
	sFlags.Parse(args)

	addr := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect", "host", host, "port", port)
		os.Exit(tuning.ExitFailure)
	}
	defer conn.Close()

	slog.Info("connected to server", "host", host, "port", port)
	c := pb.NewTunerClient(conn)

	slog.Debug("requesting epd info")
	epdInfo, err := c.RequestEPDInfo(context.Background(), &pb.EPDInfoRequest{})
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

			slog.Info(
				"downloading epd",
				"filename", epdInfo.Filename,
				"checksum", base64.URLEncoding.EncodeToString(epdInfo.Checksum))

			stream, err := c.StreamEPD(context.Background(), &pb.EPDStreamRequest{})
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

				_, err = f.WriteString(line.Line + "\n")
				if err != nil {
					slog.Warn("write error", "error", err)
				}
			}

			f.Close()
		} else {
			myChecksum, err := epdF.Checksum()
			if err != nil {
				slog.Error("checksum calculation error", "error", err)
				os.Exit(tuning.ExitFailure)
			}

			if !slices.Equal(myChecksum, epdInfo.Checksum) {
				epdF.Close()

				slog.Warn(
					"epd checksum mismatch",
					"filename", epdInfo.Filename,
					"received.Checksum", base64.URLEncoding.EncodeToString(epdInfo.Checksum),
					"local.Checksum", base64.URLEncoding.EncodeToString(myChecksum))

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

	for {
		slog.Debug("requesting job")
		r, err := c.RequestJob(context.Background(), &pb.JobRequest{})
		if err != nil {
			slog.Error("job request error", "error", err)
			continue
		}
		slog.Info("received job", "uuid", r.JobUuid)
	}
}
