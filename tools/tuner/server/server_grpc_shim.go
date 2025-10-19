package server

import (
	"context"

	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
	"github.com/paulsonkoly/chess-3/tools/tuner/shim"
	"google.golang.org/grpc"
)

type tunerServer struct {
	pb.UnimplementedTunerServer
	jobQueue    chan shim.Job
	resultQueue chan shim.Result
	epdF        *epd.File
}

func (s tunerServer) RequestEPDInfo(context.Context, *pb.EPDInfoRequest) (*pb.EPDInfo, error) {
	chkSum, err := s.epdF.Checksum()
	if err != nil {
		return nil, err
	}
	return &pb.EPDInfo{Filename: s.epdF.Basename(), Checksum: chkSum.Bytes()}, nil
}

type ShimStreamer struct {
	stream grpc.ServerStreamingServer[pb.EPDLine]
}

func (s ShimStreamer) Send(line string) error {
	return s.stream.Send(&pb.EPDLine{Line: line})
}

func (s tunerServer) StreamEPD(_ *pb.EPDStreamRequest, stream grpc.ServerStreamingServer[pb.EPDLine]) error {
	s.epdF.Stream(ShimStreamer{stream})
	return nil
}

func (s tunerServer) RequestJob(_ context.Context, _ *pb.JobRequest) (*pb.JobResponse, error) {
	job := <-s.jobQueue

	return job.ToGrpc()
}

func (s tunerServer) RegisterResult(_ context.Context, r *pb.ResultRequest) (*pb.ResultAck, error) {
	result, err := shim.ResultFromGrpc(r)
	if err != nil {
		s.resultQueue <- result
	}

	return nil, err
}
