# Wave A CI gate: go vet + test + build (hard fail).
# Full `vet check` / `gcl gate` tighten to hard failure in Wave B.

.PHONY: test vet build check-ci

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o /dev/null .

check-ci: vet test build
