package shim

import (
	"context"
	"net"

	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
	"google.golang.org/grpc"
)

type Server struct {
	grpc  *grpc.Server
	tuner tunerServer
}

func NewServer(epdF *epd.File, jobQueue <-chan Job, resultQueue chan<- Result) Server {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	s := tunerServer{
		jobQueue:    jobQueue,
		resultQueue: resultQueue,
		epdF:        epdF,
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
	epdF        *epd.File
}

func (s tunerServer) RequestEPDInfo(context.Context, *pb.EPDInfoRequest) (*pb.EPDInfo, error) {
	chkSum, err := s.epdF.Checksum()
	if err != nil {
		return nil, err
	}
	return &pb.EPDInfo{Filename: s.epdF.Basename(), Checksum: chkSum.Bytes()}, nil
}

// TODO this shouldn't be here
type streamer struct {
	stream grpc.ServerStreamingServer[pb.EPDLine]
}

func (s streamer) Send(line string) error {
	return s.stream.Send(&pb.EPDLine{Line: line})
}

func (s tunerServer) StreamEPD(_ *pb.EPDStreamRequest, stream grpc.ServerStreamingServer[pb.EPDLine]) error {
	s.epdF.Stream(streamer{stream})
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
