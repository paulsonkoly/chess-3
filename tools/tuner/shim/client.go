package shim

import (
	"context"
	"fmt"

	pb "github.com/paulsonkoly/chess-3/tools/tuner/grpc/tuner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn *grpc.ClientConn
	grpc pb.TunerClient
}

func NewClient(host string, port int) (Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return Client{}, err
	}

	return Client{conn: conn, grpc: pb.NewTunerClient(conn)}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) StreamEPD() (Stream, error) {
	gstream, err := c.grpc.StreamEPD(context.Background(), &pb.EPDStreamRequest{})
	if err != nil {
		return Stream{}, err
	}
	return Stream{grpc: gstream}, nil
}

type Stream struct {
	grpc grpc.ServerStreamingClient[pb.EPDLine]
}

func (s *Stream) Recv() (string, error) {
	gLine, err := s.grpc.Recv()
	if err != nil {
		return "", err
	}
	return gLine.Line, nil
}

func (c *Client) RequestEPDInfo() (EPDInfo, error) {
	gepdInfo, err := c.grpc.RequestEPDInfo(context.Background(), &pb.EPDInfoRequest{})
	if err != nil {
		return EPDInfo{}, nil
	}
	return epdInfoFromGrpc(gepdInfo)
}

func (c *Client) RequestJob() (Job, error) {
	gJob, err := c.grpc.RequestJob(context.Background(), &pb.JobRequest{})
	if err != nil {
		return Job{}, nil
	}

	return jobFromGrpc(gJob)
}

func (c *Client) RegisterResult(result Result) error {
	gResult, err := result.toGrpc()
	if err != nil {
		return err
	}
	_, err = c.grpc.RegisterResult(context.Background(), gResult)
	return err
}
