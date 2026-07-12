# Makefile — vet CLI release automation
#
# Fully automated flow:  make release VERSION=0.1.1
#   1. creates git tag v0.1.1
#   2. pushes the tag  ->  GitHub Action (.github/workflows/release.yml)
#                         builds 6 cross-platform binaries and creates the
#                         GitHub Release automatically (no manual steps).
#
# Config is read from .env if present (gitignored). Supported keys:
#   GITHUB_TOKEN   used only for direct `make release-api` (optional)
#   INSTALL_DIR    default install location for `make install`

-include .env

REPO        ?= buhaiqing/ve-skills
VERSION     ?= 0.1.3
TAG         := v$(VERSION)
REMOTE      ?= origin

.PHONY: help tag push release release-api build test install clean check

help:
	@echo "Targets:"
	@echo "  make release VERSION=0.1.1      tag v0.1.1 + push  -> GitHub Action publishes the release"
	@echo "  make tag VERSION=0.1.1          create + push tag only"
	@echo "  make release-api VERSION=0.1.1  create GitHub Release directly via API (needs GITHUB_TOKEN)"
	@echo "  make build                      build vet for the current platform"
	@echo "  make test                       run go tests"
	@echo "  make install                    install vet via install.sh"
	@echo "  make check                      validate .goreleaser.yaml"

# tag + push  ->  triggers the release GitHub Action (recommended, fully automated)
release: tag

tag:
	@git diff --quiet || { echo "error: working tree is dirty; commit changes first"; exit 1; }
	@git tag -f $(TAG)
	@git push $(REMOTE) $(TAG)
	@echo "pushed $(TAG) -> GitHub Action will build + publish the release"

# direct GitHub Release via API (alternative; needs GITHUB_TOKEN in .env)
release-api:
	@test -n "$(GITHUB_TOKEN)" || { echo "error: GITHUB_TOKEN not set (add to .env)"; exit 1; }
	@command -v goreleaser >/dev/null 2>&1 || { echo "error: goreleaser not installed"; exit 1; }
	goreleaser release --clean

build:
	cd cmd/vet && go build -o ../bin/vet .

test:
	cd cmd/vet && go test ./...

install:
	bash cmd/vet/install.sh

check:
	goreleaser check

clean:
	rm -rf dist cmd/bin
