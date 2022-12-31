CMD = ./cmd/russel

BIN     = ./bin
OUT     = $(BIN)/app
INSTALL = /usr/bin/russel

GO = go

compile:
	$(GO) build -o $(OUT) $(CMD)

$(BIN):
	mkdir -p $(BIN)

run:
	$(GO) run $(CMD)

install:
	cp $(OUT) $(INSTALL)

clean:
	rm -r $(BIN)/*

all:
	@echo compile, run, install, clean
