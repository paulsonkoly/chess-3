package board

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFENConversion(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{
			name: "Initial Position",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1",
		},
		{
			name: "Empty Board",
			fen:  "8/8/8/8/8/8/8/8 w - - 0 1",
		},
		{
			name: "Single Piece",
			fen:  "8/8/8/8/4N3/8/8/8 w - - 0 1",
		},
		{
			name: "Multiple Pieces",
			fen:  "rnbqkbnr/pppppppp/8/8/4N3/8/PPPPPPPP/RNBQKBNR w - - 0 1",
		},
    // TODO implement me
		// {
		// 	name: "Complex Position",
		// 	fen:  "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2",
		// },
		{
			name: "Checkmate Position",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b - - 0 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert FEN to Board
			board := FromFEN(tt.fen)

			// Convert Board back to FEN
			outputFEN := board.FEN()

			// Compare the FENs using cmp
			if diff := cmp.Diff(tt.fen, outputFEN); diff != "" {
				t.Errorf("FEN conversion failed (-expected +actual):\n%s", diff)
			}
		})
	}
}

