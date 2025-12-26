package shim

import (
	"context"
	"net"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/chess"
	"github.com/paulsonkoly/chess-3/move"
	pb "github.com/paulsonkoly/chess-3/tools/datagen/grpc/datagen"
	"google.golang.org/grpc"
)

type Server struct {
	grpc    *grpc.Server
	datagen datagenServer
}

func NewServer(config *Config, openings <-chan *board.Board, games chan<- Game) *Server {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	s := datagenServer{
		config:   config,
		openings: openings,
		games:    games,
	}
	pb.RegisterDatagenServer(grpcServer, s)
	return &Server{datagen: s, grpc: grpcServer}
}

func (s *Server) Serve(lis net.Listener) error {
	return s.grpc.Serve(lis)
}

func (s *Server) Stop() {
	s.grpc.GracefulStop()
}

type datagenServer struct {
	config   *Config
	openings <-chan *board.Board
	games    chan<- Game
	pb.UnimplementedDatagenServer
}

func (d datagenServer) RequestConfig(_ context.Context, _ *pb.ConfigRequest) (*pb.Config, error) {
	return &pb.Config{
		SoftNodes:  int64(d.config.SoftNodes),
		HardNodes:  int64(d.config.HardNodes),
		Draw:       d.config.Draw,
		DrawAfter:  int32(d.config.DrawAfter),
		DrawMargin: int32(d.config.DrawMargin),
		DrawCount:  int32(d.config.DrawCount),
		Win:        d.config.Win,
		WinAfter:   int32(d.config.WinAfter),
		WinMargin:  int32(d.config.WinMargin),
		WinCount:   int32(d.config.WinCount),
	}, nil
}

func (d datagenServer) RequestOpening(_ context.Context, _ *pb.OpeningRequest) (*pb.Opening, error) {
	b, ok := <-d.openings
	var fen string
	if ok {
		fen = b.FEN()
	}
	return &pb.Opening{Fen: fen}, nil
}

func (d datagenServer) RegisterGame(_ context.Context, gGame *pb.Game) (*pb.GameAck, error) {
	positions := make([]Position, 0, len(gGame.Positions))

	for _, position := range gGame.Positions {
		positions = append(positions,
			Position{FEN: position.Fen, BM: move.Move(position.BestMove), Score: chess.Score(position.Score)})
	}

	game := Game{
		WDL:       WDL(gGame.Wdl),
		Positions: positions,
	}

	d.games <- game
	return &pb.GameAck{}, nil
}
