#!/usr/bin/env python3
"""crash.py
Usage:
  ./crash.py <pgnfile> <white|black>
  ./crash.py <white|black> <pgnfile>

Outputs a UCI-style replay script that uses FEN from PGN (if present)
and emits `position ...` / `go nodes N` for the chosen color only.
"""
import sys
import chess
import chess.pgn
import re
from pathlib import Path

def extract_nodes_from_comment(comment):
    """
    Extract the node count from the comment.
    - Handles comments like: "+0.54 16/0 1014 958275"
    - Also handles: "0 45250, Black disconnects"
    - Skips comments containing 'book' (case-insensitive).
    - Returns an int node-count or None.
    """
    if not comment:
        return None
    if 'book' in comment.lower():
        return None

    # Make parsing robust: remove commas, convert slashes to spaces, then find integers
    clean = comment.replace("/", " ").replace(",", " ")
    ints = re.findall(r"-?\d+", clean)
    if not ints:
        return None

    # Return the last integer 
    try:
        return int(ints[-1])
    except ValueError:
        return None

def normalize_args(argv):
    """Allow either order: (pgnfile, color) or (color, pgnfile)."""
    if len(argv) != 3:
        print("Usage: ./crash.py <pgnfile> <white|black>")
        print("   or: ./crash.py <white|black> <pgnfile>")
        sys.exit(1)
    a, b = argv[1], argv[2]
    if a.lower() in ("white", "black"):
        color = a.lower()
        pgn = b
    elif b.lower() in ("white", "black"):
        color = b.lower()
        pgn = a
    else:
        print("Second argument must be 'white' or 'black' (or first).")
        sys.exit(1)
    return pgn, color

def main():
    pgn_path, color = normalize_args(sys.argv)
    want_white = (color == "white")

    pgn_file = Path(pgn_path)
    if not pgn_file.exists():
        print(f"PGN file not found: {pgn_path}")
        sys.exit(1)

    # Print UCI prelude once
    print("uci")
    # print("setoption name Hash value 4")
    print("ucinewgame")

    with open(pgn_file, "r", encoding="utf-8") as f:
        first_game = True
        while True:
            game = chess.pgn.read_game(f)
            if game is None:
                break

            # For multiple games: separate and re-init engine
            if not first_game:
                print()
                print("# --- next game ---")
                print("ucinewgame")
            first_game = False

            # Determine starting position: FEN tag preferred, else startpos
            headers = game.headers
            if "FEN" in headers:
                fen = headers["FEN"]
                # we'll not print a position line now; we print it before each go command
                starting_board = chess.Board(fen)
                use_fen = True
            else:
                starting_board = chess.Board()
                use_fen = False

            board = starting_board.copy(stack=False)  # working board
            moves = []

            # Walk moves; node is after each move
            for node in game.mainline():
                move = node.move

                if (board.turn == chess.WHITE and want_white) or (board.turn == chess.BLACK and not want_white):

                    movestr = ""
                    if moves != []:
                        movestr = " moves " + " ".join(moves)

                    nodes = extract_nodes_from_comment(node.comment)
                    if not nodes is None:
                        # Print position line using FEN if available, else startpos
                        if use_fen:
                            # If there are no moves yet, still include "moves" token with nothing after
                            print(f"position fen {fen}" + movestr)
                        else:
                            print("position startpos" + movestr)

                        print(f"go nodes {nodes}")

                board.push(move)
                moves.append(move.uci())

            movestr = " moves " + " ".join(moves)
            # the last move before the crash
            if use_fen:
                # If there are no moves yet, still include "moves" token with nothing after
                print(f"position fen {fen}" + movestr)
            else:
                print("position startpos" + movestr)
            print(f"go depth 64")



                # # Who played the move we just pushed?
                # # board.turn is the side to move *next*:
                # #   if board.turn == BLACK -> last move was WHITE
                # #   if board.turn == WHITE -> last move was BLACK
                # last_move_was_white = (board.turn == chess.BLACK)
                # last_move_was_black = (board.turn == chess.WHITE)
                #
                # # Only emit for the requested side
                # if last_move_was_white and not want_white:
                #     continue
                # if last_move_was_black and want_white == False and want_white is None:
                #     # defensive, shouldn't happen
                #     continue
                # if last_move_was_black and want_white:
                #     continue
                # if last_move_was_white and not want_white:
                #     continue
                #
                # Extract node count from comment (robust)

if __name__ == "__main__":
    main()

