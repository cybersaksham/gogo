.PHONY: test race lint build check example-blog-test docs-links docs-examples docs-generated docs-generated-update docs-tutorials docs-verify

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

docs-links:
	go run ./scripts/verify_docs.go links

docs-examples:
	go run ./scripts/verify_docs.go examples

docs-generated:
	go run ./scripts/verify_docs.go generated

docs-generated-update:
	go run ./scripts/verify_docs.go update-generated

docs-tutorials:
	go run ./scripts/verify_docs.go tutorials

docs-verify:
	go run ./scripts/verify_docs.go all
