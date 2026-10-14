# Wave A CI gate: gofmt + go vet + test + build + GCL structural gate (hard fail).

.PHONY: test vet build gofmt lint gcl-gate check-ci

gofmt:
	@test -z "$$(gofmt -l .)" || { echo "gofmt: unformatted files:"; gofmt -l .; exit 1; }

lint: gofmt vet

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o /dev/null .

gcl-gate:
	go run . gcl gate

check-ci: vet test build gcl-gate