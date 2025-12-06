package shim

import (
	"context"
	"fmt"

	"github.com/paulsonkoly/chess-3/board"
	pb "github.com/paulsonkoly/chess-3/tools/datagen/grpc/datagen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn *grpc.ClientConn
	grpc pb.DatagenClient
}

func NewClient(host string, port int) (Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return Client{}, err
	}

	return Client{conn: conn, grpc: pb.NewDatagenClient(conn)}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) RequestConfig() (Config, error) {
	gConfig, err := c.grpc.RequestConfig(context.Background(), &pb.ConfigRequest{})
	if err != nil {
		return Config{}, nil
	}
	return Config{
		SoftNodes:  int(gConfig.SoftNodes),
		HardNodes:  int(gConfig.HardNodes),
		Draw:       gConfig.Draw,
		DrawAfter:  int(gConfig.DrawAfter),
		DrawMargin: int(gConfig.DrawMargin),
		DrawCount:  int(gConfig.DrawCount),
		Win:        gConfig.Win,
		WinAfter:   int(gConfig.WinAfter),
		WinMargin:  int(gConfig.WinMargin),
		WinCount:   int(gConfig.WinCount),
	}, nil
}

func (c *Client) RequestOpening() (*board.Board, error) {
	gOpening, err := c.grpc.RequestOpening(context.Background(), &pb.OpeningRequest{})
	if err != nil {
		return nil, err
	}

	b, err := board.FromFEN(gOpening.Fen)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (c *Client) RegisterGame(g *Game) error {
	positions := make([]*pb.Position, 0, len(g.Positions))

	for _, position := range g.Positions {
		positions = append(positions,
			&pb.Position{Fen: position.FEN, BestMove: int32(position.BM), Score: int32(position.Score)})
	}

	gGame := pb.Game{
		Wdl:      int32(g.WDL),
		Positons: positions,
	}
	_, err := c.grpc.RegisterGame(context.Background(), &gGame)

	return err
}
