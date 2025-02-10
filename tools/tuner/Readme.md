# Generating a training dataset

Extract EPDs from the PGN with given filter conditions from tag.

```
$ pgn-extract -t tag -Wepd <games.pgn> > positions.epd
```

Convert EPD comments to ; \<result\> labels. Remove empty lines.

```
$ sed -f ./c0c1.sed < positions.epd | grep -v "^$" > positions-converted.epd
```

Remove non quiet or terminal positions.

```
$ ./tuner -epd positions-converted.epd -filter > positions-quiet.epd
```

Run the tuner:

```
$ ./tuner -epd positions-quiet.epd
```

## Useful tuner flags

Filtering the epd for quiet positions:

```
$ ./tuner -epd positions.epd -filter
```

The filter flag outputs positions from the EPD that doesn't have any immediate captures, checks, the side to move is not in check, and the position is not terminal: not checkmate or stalemate.

Removing positions from an EPD present in an other EPD:

```
$ ./tuner -epd fileA.epd -diff fileB.epd
```

Outputs positions from fileB.epd not present in fileA.epd. Presence check is subject to Zobrist hash collision.

Listing top 10 mis-evaluated positions:

```
$ ./tuner -epd positions.epd -misEval
```

