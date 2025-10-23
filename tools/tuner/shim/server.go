package shim

import (
	"context"
	"log/slog"
	"net"
	"path"

	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
	"google.golang.org/grpc"
)

type Server struct {
	grpc  *grpc.Server
	tuner tunerServer
}

func NewServer(fn string, jobQueue <-chan Job, resultQueue chan<- Result) Server {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	s := tunerServer{
		jobQueue:    jobQueue,
		resultQueue: resultQueue,
		filename:    fn,
	}
	pb.RegisterTunerServer(grpcServer, s)
	return Server{tuner: s, grpc: grpcServer}
}

func (s Server) Serve(lis net.Listener) {
	s.grpc.Serve(lis)
}

type tunerServer struct {
	pb.UnimplementedTunerServer
	jobQueue    <-chan Job
	resultQueue chan<- Result
	filename    string
}

func (s tunerServer) RequestEPDInfo(context.Context, *pb.EPDInfoRequest) (*pb.EPDInfo, error) {
	chkSum, err := epd.Checksum(s.filename)
	if err != nil {
		return nil, err
	}
	base := path.Base(s.filename)
	slog.Info("responding epdInfo", "Filename", base, "Checksum", chkSum)
	return &pb.EPDInfo{Filename: base, Checksum: chkSum.Bytes()}, nil
}

type streamer struct {
	stream grpc.ServerStreamingServer[pb.EPDLine]
}

func (s streamer) Send(line string) error {
	return s.stream.Send(&pb.EPDLine{Line: line})
}

func (s tunerServer) StreamEPD(_ *pb.EPDStreamRequest, stream grpc.ServerStreamingServer[pb.EPDLine]) error {
	epd.Stream(s.filename, streamer{stream})
	return nil
}

func (s tunerServer) RequestJob(_ context.Context, _ *pb.JobRequest) (*pb.JobResponse, error) {
	job := <-s.jobQueue

	return job.toGrpc()
}

func (s tunerServer) RegisterResult(_ context.Context, r *pb.ResultRequest) (*pb.ResultAck, error) {
	result, err := resultFromGrpc(r)
	if err != nil {
		return nil, err
	}

	s.resultQueue <- result
	return nil, nil
}
