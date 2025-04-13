package board_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/paulsonkoly/chess-3/board"
	"github.com/stretchr/testify/assert"
)

func TestFENConversion(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		err  error
	}{
		{
			name: "Initial Position",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1",
		},
		{ // TODO review these examples.
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
		{
			name: "Complex Position",
			fen:  "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 1",
		},
		{
			name: "Checkmate Position",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b - - 0 1",
		},
		// review up to here
		{
			name: "en passant",
			fen:  "r1bqkbnr/p1pppppp/n7/Pp6/8/8/1PPPPPPP/RNBQKBNR w - b6 0 1",
		},

		// Error Test Cases
		{
			name: "rand string",
			fen:  "adbc",
			err:  errors.New("invalid char a"),
		},
		{
			name: "Invalid character in piece placement",
			fen:  "rnbqkbnr/ppp1pppp/8/3pX3/8/8/PPPPPPPP/RNBQKBNR w KQkq d6 0 1",
			err:  errors.New("invalid char X"),
		},
		{
			name: "Overflow the board with piece placement",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e6 0 1",
			err:  errors.New("invalid position"),
		},
		{
			name: "Premature end after piece placement",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR",
			err:  errors.New("premature end of fen"),
		},
		{
			name: "Invalid active color",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq e6 0 1",
			err:  errors.New("w or b expected, got x"),
		},
		{
			name: "Invalid castling rights character",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KA e6 0 1",
			err:  errors.New("K, Q, k, q or - expected got A"),
		},
		{
			name: "En passant square invalid (a9)",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQ a9 0 1",
			err:  errors.New("square expected got a9"),
		},
		{
			name: "En passant square single character",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQ a 0 1",
			err:  errors.New("square expected got a "),
		},
		{
			name: "Fifty-move counter non-digit",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQ e6 a 1",
			err:  errors.New("digit expected got a"),
		},
		{
			name: "Premature end after active color",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w",
			err:  errors.New("premature end of fen"),
		},
		{
			name: "Premature end after castling rights",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQ",
			err:  errors.New("premature end of fen"),
		},
		{
			name: "Invalid en passant square (i3)",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQ i3 0 1",
			err:  errors.New("square expected got i3"),
		},
		{
			name: "Missing space after piece placement",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNRw KQkq e6 0 1",
			err:  errors.New("invalid char w"),
		},
		{
			name: "Invalid castling rights mix",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQx e6 0 1",
			err:  errors.New("K, Q, k, q or - expected got x"),
		},
		{
			name: "Incomplete en passant square",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQ e 0 1",
			err:  errors.New("square expected got e "),
		},
		{
			name: "Too many ranks in position",
			fen:  "8/8/8/8/8/8/8/8/8 w - - 0 1",
			err:  errors.New("invalid position"),
		},
		{
			name: "Invalid en passant file",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQ i3 0 1",
			err:  errors.New("square expected got i3"),
		},
		{
			name: "Invalid en passant rank",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQ a9 0 1",
			err:  errors.New("square expected got a9"),
		},
		{
			name: "Rank overflow",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/8/8/PPPPPPPP/RNBQKBNR w KQ - 0 1",
			err:  errors.New("invalid position"),
		},
		{
			name: "Fifty-move non-digit",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQ - x 1",
			err:  errors.New("digit expected got x"),
		},
		{
			name: "Missing space between fields",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNRwKQ-01",
			err:  errors.New("invalid char w"),
		},
		{
			name: "Invalid castling mix",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KX - 0 1",
			err:  errors.New("K, Q, k, q or - expected got X"),
		},
		{
			name: "Invalid rank structure",
			fen:  "rnbqkbnr/pppppppp/9/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1",
			err:  errors.New("invalid char 9"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert FEN to Board
			board, err := board.FromFEN(tt.fen)

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {

				assert.NoError(t, err)
				assert.NotNil(t, board)

				if board != nil {

					// Convert Board back to FEN
					outputFEN := board.FEN()

					// Compare the FENs using cmp
					if diff := cmp.Diff(tt.fen, outputFEN); diff != "" {
						t.Errorf("FEN conversion failed (-expected +actual):\n%s", diff)
					}
				}
			}
		})
	}
}
