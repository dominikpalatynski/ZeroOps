#!/usr/bin/env bash

set -e

REPO="DominikPalatynski/zeroops"
VERSION="${VERSION:-latest}"

detect_platform() {
  OS="$(uname | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "❌ Unsupported architecture: $ARCH" && exit 1 ;;
  esac

  echo "${OS}_${ARCH}"
}

PLATFORM=$(detect_platform)

echo "📦 Downloading zeroops ${VERSION} for ${PLATFORM}..."

# Get latest version if VERSION=latest
if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep tag_name | cut -d '"' -f4)
fi

FILENAME="zeroops_${VERSION#v}_${PLATFORM}"
EXT="tar.gz"
[[ "$PLATFORM" == windows_* ]] && EXT="zip"

URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}.${EXT}"

curl -sL "$URL" -o "/tmp/${FILENAME}.${EXT}"

echo "📂 Extracting..."
mkdir -p /tmp/zeroops-install
cd /tmp/zeroops-install

if [ "$EXT" = "zip" ]; then
  unzip -q "/tmp/${FILENAME}.${EXT}"
else
  tar -xzf "/tmp/${FILENAME}.${EXT}"
fi

mkdir -p "$HOME/.local/bin"
cp zeroops* "$HOME/.local/bin/zeroops"
chmod +x "$HOME/.local/bin/zeroops"

echo "✅ Installed to ~/.local/bin/zeroops"

# Ensure PATH includes ~/.local/bin
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
  SHELL_RC=""
  if [ -n "$ZSH_VERSION" ]; then
    SHELL_RC="$HOME/.zshrc"
  elif [ -n "$BASH_VERSION" ]; then
    SHELL_RC="$HOME/.bashrc"
  fi

  if [ -n "$SHELL_RC" ]; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$SHELL_RC"
    echo "✅ Added ~/.local/bin to PATH in $SHELL_RC"
    echo "📢 Run: source $SHELL_RC or restart terminal"
  else
    echo "⚠️ Couldn't detect shell config file. Add ~/.local/bin to your PATH manually if needed."
  fi
fi
