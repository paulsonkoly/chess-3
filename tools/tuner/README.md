# Tuner

`tuner` is the HCE coefficient tuner for chess-3. It is a client-server program, where the server orchestrates the tuning process, while the clients are performing gradient calculations on chunks of the given epd file.

## EPD

epd is a tuning data set, with each line comprising of a position epd, and a 0.0, 0.5 or 1.0 WDL label separated by `;`.

For instance:

```
5r2/p4pk1/2pb4/8/1p2rN2/4p3/PPPB4/3K4 w - - 0 3; 0.0
r2q1rk1/3n1p2/2pp3p/1pb1p1p1/p3P3/P1NP1N1P/RPP2PP1/5QK1 b - - 0 2; 0.0
rn2r2k/p1R4p/4bp2/8/1Q6/6P1/1P3P1P/6K1 w - - 0 1; 0.0
```

WDL is side relative: 0.0 means black wins.

## Usage

The same binary is used for client and server.

```
tuner [tuner flags...] [client|server] [command flags]
Usage of ./tuner:
  -h	help
  -log int
    	log level. Lower numbers enable more logs. see https://pkg.go.dev/log/slog#Level
```

### Client

The client connects to the server, downloads the epd file, and runs a number of threads concurrent on the chunks of the epd file, given by the server.

```
Usage of client:
  -host string
    	host to connect to (default "localhost")
  -port int
    	port to connect to (default 9001)
  -threads int
    	number of worker threads (default 8)
```

### Server

The server orchestrates the tuning and outputs the coefficients after each passed tuning epoch.

```
Usage of server:
  -epd string
    	epd file name
  -host string
    	host to listen on (default "localhost")
  -port int
    	port to listen on (default 9001)
```
