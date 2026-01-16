# crash

`crash` is a simple python script to reproduce chess-3 crashes on OpenBench.

Assuming we have a _PGN_ with node counts we should be able to replay the game in chess-3 and re-create crashes. `crash` takes a _PGN_ and outputs a stream of _UCI_ commands that recreates the original run.

## Caveats

Deterministic engine behaviour can only be guaranteed in a single threaded scenario. The transposition table size has to be set to the same size as the original run.

## Example

`crash.pgn`

```
[Event "?"]
[Site "?"]
[Date "2026.01.12"]
[Round "1"]
[White "chess-3-dev"]
[Black "chess-3-base"]
[Result "0-1"]
[ECO "B22"]
[GameDuration "00:00:23"]
[GameEndTime "2026-01-12T14:06:06.836 CET"]
[GameStartTime "2026-01-12T14:05:43.028 CET"]
[Opening "Sicilian"]
[PlyCount "82"]
[Termination "abandoned"]
[TimeControl "10.44+0.1"]
[Variation "Alapin's Variation (2.c3)"]

1. e4 {book} c5 {book} 2. c3 {book} d5 {book} 3. exd5 {book} Qxd5 {book}
4. d4 {book} g6 {book} 5. Nf3 {book} Nc6 {book} 6. Be2 {book} Nh6 {book}
7. c4 {book} Qd6 {book} 8. d5 {book} Ne5 {book} 9. Nxe5 {+0.35 18/0 452 731370}
Qxe5 {-0.48 19/0 441 735948} 10. O-O {+0.48 20/0 831 1345169}
Bg7 {-0.55 20/0 1132 1855138} 11. Nc3 {+0.60 18/0 579 909482}
O-O {-0.41 19/0 533 881861} 12. Re1 {+0.49 19/0 572 935855}
Qc7 {-0.41 19/0 1379 2290227} 13. Bg5 {+0.53 17/0 383 637618}
Nf5 {-0.57 16/0 359 608708} 14. Bd3 {+0.44 17/0 615 1003605}

[...]

Qxa2 {-2.32 18/0 145 254170} 38. Qxf7 {+2.22 19/0 172 296602}
Qb2 {-2.47 17/0 132 236296} 39. Qf8+ {+2.70 17/0 143 249969}
Qg7 {-2.85 18/0 270 473591} 40. Qd6 {+3.31 18/0 203 348000}
Qc3 {-2.79 18/0 138 247892} 41. h4 {+5.37 20/0 225 396155}
Qb2 {-3.66 18/0 210 375320, White disconnects} 0-1
```

Since it's white that crashed, we want to re-create the game from white's perspective:

```
$ bin/python3 crash white crash.pgn
uci
ucinewgame
position startpos moves e2e4 c7c5 c2c3 d7d5 e4d5 d8d5 d2d4 g7g6 g1f3 b8c6 f1e2 g8h6 c3c4 d5d6 d4d5 c6e5
go nodes 731370
position startpos moves e2e4 c7c5 c2c3 d7d5 e4d5 d8d5 d2d4 g7g6 g1f3 b8c6 f1e2 g8h6 c3c4 d5d6 d4d5 c6e5 f3e5 d6e5
go nodes 1345169
position startpos moves e2e4 c7c5 c2c3 d7d5 e4d5 d8d5 d2d4 g7g6 g1f3 b8c6 f1e2 g8h6 c3c4 d5d6 d4d5 c6e5 f3e5 d6e5 e1g1 f8g7
go nodes 909482
position startpos moves e2e4 c7c5 c2c3 d7d5 e4d5 d8d5 d2d4 g7g6 g1f3 b8c6 f1e2 g8h6 c3c4 d5d6 d4d5 c6e5 f3e5 d6e5 e1g1 f8g7 b1c3 e8g8
go nodes 935855
position startpos moves e2e4 c7c5 c2c3 d7d5 e4d5 d8d5 d2d4 g7g6 g1f3 b8c6 f1e2 g8h6 c3c4 d5d6 d4d5 c6e5 f3e5 d6e5 e1g1 f8g7 b1c3 e8g8 f1e1 e5c7
go nodes 637618
position startpos moves e2e4 c7c5 c2c3 d7d5 e4d5 d8d5 d2d4 g7g6 g1f3 b8c6 f1e2 g8h6 c3c4 d5d6 d4d5 c6e5 f3e5 d6e5 e1g1 f8g7 b1c3 e8g8 f1e1 e5c7 c1g5 h6f5

[...]
```

Edit this output adding the right hash setting, after the `uci` line for example:

```
setoption name Hash value 1
```

Feeding the `uci` stream on the stdin of `chess-3` should reproduce the crash.
