EXE=chess3
# only enable SPSA for SPSA tuning.
SPSA ?= 0

GIT_VERSION := $(shell git describe --tags --always --dirty)

LDFLAGS := -X github.com/paulsonkoly/chess-3/uci.GitVersion=$(GIT_VERSION)

files := $(shell find . -name '*.go')

# Optional build tags
ifeq ($(SPSA),1)
	GO_TAGS := -tags spsa
else
	GO_TAGS :=
endif

$(EXE): chess3.pprof $(files)
	go build $(GO_TAGS) -ldflags "$(LDFLAGS)" -pgo chess3.pprof -o $@ main.go

chess3.pprof: $(EXE).nopgo $(files)
	./$(EXE).nopgo -cpuProf $@ bench

$(EXE).nopgo: $(files)
	go build $(GO_TAGS) -ldflags "$(LDFLAGS)" -o $@ main.go
