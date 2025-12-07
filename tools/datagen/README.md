# datagen

A command-line tool to generate games played by chess-3. Games are saved in an Sqlite3 database for further processing. The program is a distributed client-server architecture.

## server

```
Usage of server:
  -WinAfter int
      enables win adjudication after this many moves (default 40)
  -dbFile string
      file name for the database (default "datagen.db")
  -draw
      enable draw adjudication (default true)
  -drawAfter int
      enables draw adjudication after this many moves (default 40)
  -drawCount int
      number of positions drawn back to back for adjudication (default 4)
  -drawMargin int
      position considered draw with this margin in adjudication (cp) (default 20)
  -gameCount int
      number of games to generate (default 1000000)
  -hardNodes int
      hard node count for search (default 8000000)
  -host string
      host to listen on (default "localhost")
  -openingDepth int
      number of random generated opening moves (default 8)
  -openingMargin int
      margin for what's considered to be balanced opening (cp) (default 300)
  -port int
      port to listen on (default 9001)
  -softNodes int
      soft node count for search (default 15000)
  -win
      enable win adjudication (default true)
  -winCount int
      number of positions won back to back for adjudication (default 4)
  -winMargin int
      positions considered win with this margin in adjudication (cp) (default 600)
```

## client

```
Usage of client:
  -host string
      host to connect to (default "localhost")
  -port int
      port to connect to (default 9001)
  -threads int
      number of worker threads (default 8)
```

