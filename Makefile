.PHONY: test race race-concurrency lint fmt-check build check integration deps vuln-check ci example-blog-test docs-links docs-examples docs-generated docs-generated-update docs-tutorials docs-verify

test:
	go test ./...

race:
	go test -race ./...

race-concurrency:
	go test -race ./app ./http ./queue/... ./sessions ./cache ./signals ./internal/cli

lint:
	go vet ./...

fmt-check:
	test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './.git/*'))"

build:
	go build -o bin/gogo ./cmd/gogo

check:
	go test ./...
	go vet ./...

integration:
	go test -tags=integration ./internal/cli

deps:
	go list -m all

vuln-check:
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed; skipping local vulnerability scan"; \
	fi

ci: fmt-check lint test integration race-concurrency example-blog-test docs-verify deps vuln-check

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
