.PHONY: test race lint build check example-blog-test

test:
	go test ./...

race:
	go test -race ./...

lint:
	go vet ./...

build:
	go build -o bin/gogo ./cmd/gogo

check:
	go test ./...
	go vet ./...

example-blog-test:
	go test ./examples/blog/...
