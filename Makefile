DOCS_NODE ?= npx -y node@22.12.0

.PHONY: test race race-concurrency lint fmt-check build check integration deps vuln-check bench ci example-blog-test docs-public-install docs-public-audit docs-public-check docs-public-build docs-links docs-examples docs-generated docs-generated-update docs-tutorials docs-verify

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

bench:
	go test -run '^$$' -bench . -benchmem ./benchmarks

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

docs-public-install:
	npm --prefix docs/public install

docs-public-audit:
	npm --prefix docs/public audit --audit-level=moderate

docs-public-check:
	cd docs/public && $(DOCS_NODE) node_modules/.bin/astro check

docs-public-build:
	cd docs/public && $(DOCS_NODE) node_modules/.bin/astro build

docs-tutorials:
	go run ./scripts/verify_docs.go tutorials

docs-verify:
	go run ./scripts/verify_docs.go all
	$(MAKE) docs-public-audit
	$(MAKE) docs-public-check
	$(MAKE) docs-public-build
