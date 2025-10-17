module github.com/paulsonkoly/chess-3/tools/tuner

go 1.25.0

require (
	github.com/google/uuid v1.6.0
	github.com/paulsonkoly/chess-3 v0.0.0
	golang.org/x/sys v0.34.0
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

replace github.com/paulsonkoly/chess-3 => ../../

require (
	golang.org/x/exp v0.0.0-20250218142911-aa4b98e5adaa // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
)
