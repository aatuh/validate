SHELL := /bin/bash

PKG ?= ./...
COVERAGE_OUT ?= coverage.out
GOVULNCHECK ?= $(shell go env GOPATH)/bin/govulncheck

.PHONY: tidy vet test examples race-cover coverage fuzz vuln ci finalize clean

tidy:
	go mod tidy
	git diff --exit-code -- go.mod go.sum
	test -z "$$(git status --short -- go.mod go.sum)" || (git status --short -- go.mod go.sum; exit 1)

vet:
	go vet ./...

test:
	go test "$(PKG)"

examples:
	go test ./examples -v -count 1

race-cover:
	go test ./... -race -covermode=atomic -coverprofile="$(COVERAGE_OUT)"

coverage: race-cover
	go tool cover -func="$(COVERAGE_OUT)"

fuzz:
	bash scripts/fuzz.sh

vuln:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	"$(GOVULNCHECK)" ./...

ci: tidy vet test examples vuln coverage fuzz

finalize: ci

clean:
	rm -f "$(COVERAGE_OUT)"
