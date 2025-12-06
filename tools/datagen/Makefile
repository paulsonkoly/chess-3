GOSRC=$(shell find . -name '*.go')

datagen: $(GOSRC)
	go build -o $@

DB ?= "datagen.db"

.PHONY: database
database:
	rm -f $(DB)
	sqlite3 $(DB) < sql/create_db.sql

.PHONY: grpc
grpc: protos/datagen.proto
	protoc --go_out=. --go-grpc_out=. $<

