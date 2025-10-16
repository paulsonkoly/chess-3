package server

import (
	"context"

	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
)

type tunerServer struct {
	pb.UnimplementedTunerServer
	jobQueue    chan Job
	resultQueue chan Result
	epdFilename string
	epdChecksum []byte
}

func (s tunerServer) RequestEPDInfo(context.Context, *pb.EPDInfoRequest) (*pb.EPDInfo, error) {
	return &pb.EPDInfo{Filename: s.epdFilename, Checksum: s.epdChecksum}, nil
}

func (s tunerServer) RequestJob(_ context.Context, _ *pb.JobRequest) (*pb.JobResponse, error) {
	job := <-s.jobQueue

	result := pb.JobResponse{
		JobUuid: job.UUID,
	}

	return &result, nil
}

func (s tunerServer) RegisterResult(_ context.Context, r *pb.ResultRequest) (*pb.ResultAck, error) {
	result := Result{UUID: r.JobUuid}
	s.resultQueue <- result

	return nil, nil
}
