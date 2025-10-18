package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
	"google.golang.org/grpc"
)

type tunerServer struct {
	pb.UnimplementedTunerServer
	jobQueue    chan queueJob
	resultQueue chan result
	epdF        *epd.File
}

func (s tunerServer) RequestEPDInfo(context.Context, *pb.EPDInfoRequest) (*pb.EPDInfo, error) {
	chkSum, err := s.epdF.Checksum()
	if err != nil {
		return nil, err
	}
	return &pb.EPDInfo{Filename: s.epdF.Basename(), Checksum: chkSum}, nil
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

	uuidBytes, err := job.uuid.MarshalBinary()
	if err != nil {
		return nil, err
	}

	result := pb.JobResponse{
		Uuid:         uuidBytes,
		Epoch:        int32(job.epoch),
		Start:        int32(job.start),
		End:          int32(job.end),
		Checksum:     job.checksum,
		Coefficients: job.coefficients,
		K:            job.k,
	}

	return &result, nil
}

func (s tunerServer) RegisterResult(_ context.Context, r *pb.ResultRequest) (*pb.ResultAck, error) {
	uuid, err := uuid.FromBytes(r.Uuid)
	if err != nil {
		return nil, err
	}
	s.resultQueue <- result{uuid: uuid}

	return nil, nil
}
