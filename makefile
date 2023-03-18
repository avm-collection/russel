CMD = ./cmd/russel

BIN     = bin
OUT     = $(BIN)/app
INSTALL = /usr/bin/russel

ERROR_TESTS = tests/no_entry_error.rsl tests/errors.rsl tests/name_suggest.rsl
TESTS       = $(filter-out $(ERROR_TESTS),$(wildcard tests/*.rsl))
BIN_TESTS   = $(subst tests/,$(BIN)/,$(basename $(TESTS)))

GO = go

.PHONY: run

compile:
	$(GO) build -o $(OUT) $(CMD)

$(BIN):
	mkdir -p $(BIN)

run:
	$(GO) run $(CMD)

tests: $(INSTALL) $(BIN_TESTS)

$(BIN_TESTS): $(BIN)/% : tests/%.rsl
	russel build $< -o $@

install:
	cp $(OUT) $(INSTALL)

clean:
	rm -r $(BIN)/*

all:
	@echo compile, run, install, clean
