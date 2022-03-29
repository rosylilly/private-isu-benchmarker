.PHONY: build
build: bin/benchmarker

.PHONY: test
test:
	go test ./...

.PHONY: run
run: build
	./bin/benchmarker

bin/benchmarker: $(shell find . -name '*.go' -print)
	go build -o $@ .
