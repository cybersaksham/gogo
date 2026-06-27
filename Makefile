.PHONY: test race lint build check

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
