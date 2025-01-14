package debug

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/movegen"
	"github.com/paulsonkoly/chess-3/mstore"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

func Perft(b *board.Board, depth int) int {
	ms := mstore.New()
	return perft(ms, b, depth)
}

func perft(ms *mstore.MStore, b *board.Board, depth int) int {
	if depth == 0 {
		return 1
	}

	perft := 0
	me := b.STM

	ms.Push()
	defer ms.Pop()

	movegen.GenMoves(ms, b, board.Full)

	for _, m := range ms.Frame() {
		b.MakeMove(&m)

		if b.Hashes[len(b.Hashes)-1] != b.Hash() {
			panic("oops")
		}

		kingBB := b.Pieces[King] & b.Colors[me]
		if !movegen.IsAttacked(b, me.Flip(), kingBB) {
			perft += Perft(b, depth-1)
		}

		b.UndoMove(&m)
	}

	return perft
}

func StockfishPerft(b *board.Board, depth int) int {
	var result int
	sf := exec.Command("stockfish")
	sfIn, err := sf.StdinPipe()
	if err != nil {
		panic(err)
	}
	sfOut, err := sf.StdoutPipe()
	if err != nil {
		panic(err)
	}

	err = sf.Start()
	if err != nil {
		panic(err)
	}

	_, err = sfIn.Write([]byte(fmt.Sprintf("position fen %s\n", b.FEN())))
	if err != nil {
		panic(err)
	}

	_, err = sfIn.Write([]byte(fmt.Sprintf("go perft %d\n", depth)))
	if err != nil {
		panic(err)
	}

	sfIn.Close()

	reader := bufio.NewReader(sfOut)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break // Exit loop on EOF or error
		}
		if strings.HasPrefix(line, "Nodes searched: ") {
			line := strings.TrimSpace(line)
			result, err = strconv.Atoi(strings.TrimPrefix(line, "Nodes searched: "))
			if err != nil {
				panic(err)
			}
			break
		}
	}

	err = sf.Wait()
	if err != nil {
		panic(err)
	}
	return result
}

func MatchPerft(b *board.Board, depth int) {
	ms := mstore.New()

	matchPerft(ms, b, depth)
}

func matchPerft(ms *mstore.MStore, b *board.Board, depth int) {
	if depth <= 0 {
		return
	}

	sfs, own := StockfishPerft(b, depth), perft(ms, b, depth)

	if own != sfs {
		fmt.Printf("%s at depth %d stockfish %d own %d\n", b.FEN(), depth, sfs, own)

		ms.Push()
		defer ms.Pop()

		movegen.GenMoves(ms, b, board.Full)

		me := b.STM
		for _, m := range ms.Frame() {
			b.MakeMove(&m)

			kingBB := b.Pieces[King] & b.Colors[me]
			if !movegen.IsAttacked(b, me.Flip(), kingBB) {
				MatchPerft(b, depth-1)
			}
			b.UndoMove(&m)
		}
	}
}
