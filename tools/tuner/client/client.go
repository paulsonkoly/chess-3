package client

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
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

	addr:=fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect", "host", host, "port", port)
	}
	defer conn.Close()

	slog.Info("connected to server", "host", host, "port", port)

	c := pb.NewTunerClient(conn)
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
