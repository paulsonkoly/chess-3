package shim

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"path"
	"time"

	"github.com/paulsonkoly/chess-3/tools/tuner/epd"
	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type Server struct {
	grpc  *grpc.Server
	tuner tunerServer
}

const connEstablishTimeout = 5 * time.Second

func NewServer(fn string, jobQueue <-chan Job, resultQueue chan<- Result, done <-chan struct{}) *Server {
	grpcServer := grpc.NewServer(grpc.ConnectionTimeout(connEstablishTimeout))
	s := tunerServer{
		jobQueue:    jobQueue,
		resultQueue: resultQueue,
		done:        done,
		filename:    fn,
	}
	pb.RegisterTunerServer(grpcServer, s)
	return &Server{tuner: s, grpc: grpcServer}
}

func (s *Server) Serve(lis net.Listener) error {
	if err := s.grpc.Serve(lis); err != nil {
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return err
	}
	return nil
}

func (s *Server) GracefulStop() {
	s.grpc.GracefulStop()
}

func (s *Server) Stop() {
	s.grpc.Stop()
}

type tunerServer struct {
	pb.UnimplementedTunerServer
	jobQueue    <-chan Job
	resultQueue chan<- Result
	done        <-chan struct{}
	filename    string
}

func (s tunerServer) RequestEPDInfo(context.Context, *pb.EPDInfoRequest) (*pb.EPDInfo, error) {
	chkSum, err := epd.Checksum(s.filename)
	if err != nil {
		return nil, err
	}
	base := path.Base(s.filename)
	slog.Debug("responding epdInfo", slog.String("filename", base), slog.String("checksum", chkSum.String()))
	return &pb.EPDInfo{Filename: base, Checksum: chkSum.Bytes()}, nil
}

type streamer struct {
	stream grpc.ServerStreamingServer[pb.EPDLine]
}

func (s streamer) Send(line string) error {
	return s.stream.Send(&pb.EPDLine{Line: line})
}

func (s tunerServer) StreamEPD(_ *pb.EPDStreamRequest, stream grpc.ServerStreamingServer[pb.EPDLine]) error {
	return epd.Stream(s.filename, streamer{stream})
}

func (s tunerServer) RequestJob(ctx context.Context, _ *pb.JobRequest) (*pb.JobResponse, error) {
	// Prefer telling the client to stop over handing out more work.
	select {
	case <-s.done:
		return &pb.JobResponse{Shutdown: true}, nil
	default:
	}

	select {
	case job := <-s.jobQueue:
		return job.toGrpc()
	case <-s.done:
		return &pb.JobResponse{Shutdown: true}, nil
	case <-ctx.Done():
		// The client went away, the server will re-schedule this job.
		return nil, status.FromContextError(ctx.Err()).Err()
	}
}

func (s tunerServer) RegisterResult(ctx context.Context, r *pb.ResultRequest) (*pb.ResultAck, error) {
	result, err := resultFromGrpc(r)
	if err != nil {
		return nil, err
	}

	select {
	case s.resultQueue <- result:
	case <-ctx.Done():
		// The client went away, the server will re-schedule this job.
		return nil, status.FromContextError(ctx.Err()).Err()
	case <-s.done:
		// The server is shutting down, the result is no longer needed. This is
		// not a client error.
	}

	return &pb.ResultAck{}, nil
}
