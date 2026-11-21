GO      := go
SERVER  := server
# sorgenti: se uno cambia, il binario viene ricompilato
SOURCES := $(shell find cmd internal -name '*.go' 2>/dev/null) go.mod

# argomenti per run-server, es: make run-server ARGS="-port 5000 -seed 42"
ARGS    ?=

all: $(SERVER)

$(SERVER): $(SOURCES)
	$(GO) build -o $(SERVER) ./cmd/server

# scarica le dipendenze del go.mod (per ora nessuna)
install:
	$(GO) mod download

run-server: $(SERVER)
	./$(SERVER) $(ARGS)

# go vet + controllo formattazione: fallisce se gofmt trova file da sistemare
lint:
	$(GO) vet ./...
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "file da formattare (make fmt):"; echo "$$out"; exit 1; \
	fi

# formatta tutto il codice
fmt:
	gofmt -w .

clean:
	rm -f $(SERVER)

re: clean all

.PHONY: all install run-server lint fmt clean re