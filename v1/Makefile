GO_FILES:=$(shell find . -type f -name '*.go' -print)

.PHONY: build
build: bin/benchmarker

.PHONY: run
run: build
	@bin/benchmarker -debug

.PHONY: clean
clean:
	@$(RM) -rf bin

bin/benchmarker: ${GO_FILES}
	@mkdir -p $(@D)
	@go build -o $@
