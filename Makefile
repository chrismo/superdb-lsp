SUPER_VERSION := v0.2.0
SUPER_REPO := https://github.com/brimdata/super.git
DEPS_DIR := _deps

.PHONY: deps build test generate clean

deps: $(DEPS_DIR)/super

$(DEPS_DIR)/super:
	@mkdir -p $(DEPS_DIR)
	git clone --branch $(SUPER_VERSION) --depth 1 $(SUPER_REPO) $(DEPS_DIR)/super

build: deps
	cd lsp && go build -v ./...

test: deps
	cd lsp && go build -v && go test -v

generate: deps
	cd lsp && go generate ./...

clean:
	rm -rf $(DEPS_DIR)
