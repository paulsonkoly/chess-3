grpc: protos/tuner.proto
	protoc --go_out=. --go-grpc_out=. $<

files=$(shell find . -name '*.go')

tuner: $(files)
	go build -o $@ $@.go
