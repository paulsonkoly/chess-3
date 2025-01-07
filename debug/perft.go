package debug

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/paulsonkoly/chess-3/board"
	"github.com/paulsonkoly/chess-3/movegen"

	//revive:disable-next-line
	. "github.com/paulsonkoly/chess-3/types"
)

func Perft(b *board.Board, depth int) int {
	if depth == 0 {
		return 1
	}

	perft := 0
	me := b.STM
	for m := range movegen.Moves(b, board.Full) {
		b.MakeMove(&m)

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
	if depth <= 0 {
		return
	}

	sfs, own := StockfishPerft(b, depth), Perft(b, depth)

	if own != sfs {
		fmt.Printf("%s at depth %d stockfish %d own %d\n", b.FEN(), depth, sfs, own)

		me := b.STM
		for m := range movegen.Moves(b, board.Full) {
			b.MakeMove(&m)

			kingBB := b.Pieces[King] & b.Colors[me]
			if !movegen.IsAttacked(b, me.Flip(), kingBB) {
				fmt.Println(b.FEN(), m)
				MatchPerft(b, depth-1)
			}
			b.UndoMove(&m)
		}
	}
}
