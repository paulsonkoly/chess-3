package client

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
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

	for haveEPD := false; !haveEPD; {
		epd, err := epd.Load(epdInfo.Filename)
		if err != nil {
			if !errors.Is(err, unix.ENOENT) {
				slog.Error("unexpected error on epd load", "error", err)
				os.Exit(tuning.ExitFailure)
			}

			slog.Info(
				"downloading epd",
				"filename", epdInfo.Filename,
				"checksum", base64.URLEncoding.EncodeToString(epdInfo.Checksum))
		} else {
			if !slices.Equal(epd.Checksum, epdInfo.Checksum) {
				slog.Warn(
					"epd checksum mismatch",
					"filename", epdInfo.Filename,
					"received.Checksum", epdInfo.Checksum,
					"local.Checksum", epd.Checksum)
				slog.Debug("deleting local epd", "filename", epd.Basename())
				if err := os.Remove(epd.Basename()); err != nil {
					slog.Error("can't remove bad epd file", "error", err)
					os.Exit(tuning.ExitFailure)
				}
			} else {
				haveEPD = true
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
