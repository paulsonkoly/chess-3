module github.com/paulsonkoly/chess-3/tools/tuner

go 1.26.0

require (
	github.com/google/uuid v1.6.0
	github.com/paulsonkoly/chess-3 v0.0.0
	github.com/prometheus/client_golang v1.25.0
	golang.org/x/sync v0.23.0
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.12
)

replace github.com/paulsonkoly/chess-3 => ../../

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	golang.org/x/exp v0.0.0-20250218142911-aa4b98e5adaa // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
)
