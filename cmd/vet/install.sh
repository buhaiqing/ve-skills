#!/usr/bin/env bash
#
# install.sh — install or update the `vet` CLI.
#
# Always fetches the LATEST GitHub Release of the ve-skills `vet` tool and
# installs the binary matching the current OS/architecture into INSTALL_DIR
# (default: /usr/local/bin). Re-running this script updates to the newest
# version (idempotent, one-shot, no background daemon).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/buhaiqing/ve-skills/main/cmd/vet/install.sh | bash
#   # or after cloning:
#   bash cmd/vet/install.sh
#
set -euo pipefail

REPO="buhaiqing/ve-skills"
BINARY="vet"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION_FILE="${INSTALL_DIR}/${BINARY}.version"

err() { echo "error: $*" >&2; exit 1; }

# --- detect OS / arch -------------------------------------------------------
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64 | amd64) ARCH="amd64" ;;
  arm64 | aarch64) ARCH="arm64" ;;
  *) err "unsupported architecture: $ARCH" ;;
esac
case "$OS" in
  linux) OS="linux" ;;
  darwin) OS="darwin" ;;
  *) err "unsupported OS: $OS (only linux/darwin supported)" ;;
esac

# --- resolve latest version -------------------------------------------------
command -v curl >/dev/null 2>&1 || err "curl is required"
echo "resolving latest $BINARY release for ${OS}/${ARCH} ..."
LATEST="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep -m1 '"tag_name"' \
  | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
[ -n "$LATEST" ] || err "could not determine latest release tag"
echo "latest: $LATEST"

# strip leading "vet/" prefix if present (tag is vet/vX.Y.Z)
VERSION="${LATEST#vet/}"

# --- find asset for this platform ------------------------------------------
ASSET="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "downloading $URL"
if ! curl -fsSL "$URL" -o "$TMP/$ASSET"; then
  # windows ships a .zip instead of .tar.gz
  ASSET="${BINARY}_${VERSION}_${OS}_${ARCH}.zip"
  URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"
  curl -fsSL "$URL" -o "$TMP/$ASSET" || err "download failed: $URL"
fi

# --- extract & install ------------------------------------------------------
if [[ "$ASSET" == *.zip ]]; then
  command -v unzip >/dev/null 2>&1 || err "unzip is required for windows asset"
  unzip -o "$TMP/$ASSET" -d "$TMP" >/dev/null
else
  tar -xzf "$TMP/$ASSET" -C "$TMP"
fi

# goreleaser archives nest the binary under the archive name dir; find it
SRC="$(find "$TMP" -type f -name "$BINARY" | head -1)"
[ -n "$SRC" ] || err "binary $BINARY not found in archive"

echo "installing to $INSTALL_DIR/$BINARY"
if [ -w "$INSTALL_DIR" ]; then
  install -m 0755 "$SRC" "$INSTALL_DIR/$BINARY"
else
  sudo install -m 0755 "$SRC" "$INSTALL_DIR/$BINARY"
fi
echo "$VERSION" > "$VERSION_FILE"

echo "✅ $BINARY $VERSION installed at $INSTALL_DIR/$BINARY"
"$INSTALL_DIR/$BINARY" version
