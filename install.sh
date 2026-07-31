#!/bin/sh
# install.sh — install aimap from the latest GitHub release
# Usage: curl -sfL https://raw.githubusercontent.com/pgulb/aimap/main/install.sh | sh

set -e

REPO="pgulb/aimap"
BINARY="aimap"

# Detect OS and architecture.
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

ARCHIVE="aimap_${OS}_${ARCH}.tar.gz"

# Fetch the latest release tag from GitHub API.
echo "Fetching latest release..."
LATEST=$(curl -sfL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$LATEST" ]; then
  echo "Failed to determine latest release."
  exit 1
fi

echo "Latest release: $LATEST"

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST/$ARCHIVE"

# Download to a temp directory.
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading $ARCHIVE..."
curl -sfL "$DOWNLOAD_URL" -o "$TMPDIR/$ARCHIVE"

echo "Extracting..."
tar xzf "$TMPDIR/$ARCHIVE" -C "$TMPDIR"

# Determine install location.
if [ -w "/usr/local/bin" ]; then
  DEST="/usr/local/bin"
elif [ -w "$HOME/.local/bin" ]; then
  DEST="$HOME/.local/bin"
else
  mkdir -p "$HOME/.local/bin"
  DEST="$HOME/.local/bin"
fi

cp "$TMPDIR/$BINARY" "$DEST/$BINARY"
chmod +x "$DEST/$BINARY"

echo ""
echo "Installed to $DEST/$BINARY"

# Check if destination is on PATH.
case ":$PATH:" in
  *":$DEST:"*) ;;
  *)
    echo ""
    echo "Warning: $DEST is not in your PATH."
    echo "Add it by running:"
    echo "  export PATH=\"\$PATH:$DEST\""
    echo "Or add that line to your shell profile (~/.bashrc, ~/.zshrc, etc.)."
    ;;
esac

echo ""
echo "Installation complete. Run 'aimap' to get started."
