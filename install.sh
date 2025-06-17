#!/usr/bin/env bash

set -e

REPO="DominikPalatynski/zeroops"
VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Normalize architecture name
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "📦 Downloading zeroops v$VERSION for $OS-$ARCH..."

FILENAME="zeroops_${VERSION}_${OS}_${ARCH}"
ARCHIVE="$FILENAME.tar.gz"
[ "$OS" = "windows" ] && ARCHIVE="$FILENAME.zip"

URL="https://github.com/$REPO/releases/download/v$VERSION/$ARCHIVE"
curl -LO "$URL"

echo "📂 Extracting..."
if [[ "$ARCHIVE" == *.zip ]]; then
  unzip -o "$ARCHIVE"
else
  tar -xzf "$ARCHIVE"
fi

BIN_DIR="/usr/local/bin"
BIN="zeroops"
[ "$OS" = "windows" ] && BIN="zeroops.exe"

if [ -w "$BIN_DIR" ]; then
  mv "$BIN" "$BIN_DIR/"
  echo "✅ Installed to $BIN_DIR/$BIN"
else
  mkdir -p ~/.local/bin
  mv "$BIN" ~/.local/bin/
  echo "⚠️ No sudo, installed to ~/.local/bin/$BIN (add it to your PATH if needed)"
fi

# Cleanup
rm -f "$ARCHIVE"
