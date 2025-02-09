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


