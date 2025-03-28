package debug

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/move"
	"github.com/paulsonkoly/chess-3/movegen"
)

func Perft(b *board.Board, depth int) int {
	ms := move.NewStore()
	return perft(ms, b, depth)
}

func perft(ms *move.Store, b *board.Board, depth int) int {
	if depth == 0 {
		return 1
	}

	if movegen.IsCheckmate(b) || movegen.IsStalemate(b) {
		return 0
	}

	cnt := 0
	me := b.STM

	ms.Push()
	defer ms.Pop()

	movegen.GenMoves(ms, b)

	hasLegal := false

	for _, m := range ms.Frame() {
		b.MakeMove(&m)

		if b.Hash() != b.CalculateHash() {
			panic("oops")
		}

		if !movegen.InCheck(b, me) {
			hasLegal = true
			cnt += perft(ms, b, depth-1)
		}

		b.UndoMove(&m)
	}

	if !hasLegal {
		panic("oops")
	}

	return cnt
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

	_, err = sfIn.Write(fmt.Appendf(nil, "position fen %s\n", b.FEN()))
	if err != nil {
		panic(err)
	}

	_, err = sfIn.Write(fmt.Appendf(nil, "go perft %d\n", depth))
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
	ms := move.NewStore()

	matchPerft(ms, b, depth)
}

func matchPerft(ms *move.Store, b *board.Board, depth int) {
	if depth <= 0 {
		return
	}

	sfs, own := StockfishPerft(b, depth), perft(ms, b, depth)

	if own != sfs {
		fmt.Printf("%s at depth %d stockfish %d own %d\n", b.FEN(), depth, sfs, own)

		ms.Push()
		defer ms.Pop()

		movegen.GenMoves(ms, b)

		me := b.STM
		for _, m := range ms.Frame() {
			b.MakeMove(&m)

			if !movegen.InCheck(b, me) {
				MatchPerft(b, depth-1)
			}
			b.UndoMove(&m)
		}
	}
}
