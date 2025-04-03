EXE=chess3

files=$(wildcard **/*.go)

$(EXE): chess3.pprof $(files)
	go build -pgo chess3.pprof -o $@ main.go

chess3.pprof: $(EXE).nopgo
	./$(EXE).nopgo -cpuProf $@ bench 

$(EXE).nopgo: $(files)
	go build -o $@ main.go
