datagen: datagen.go
	go build -o $@ $<

DB ?= "datagen.db"

.PHONY: database
database:
	rm -f $(DB)
	sqlite3 $(DB) < sql/create_db.sql
