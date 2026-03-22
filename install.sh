#!/bin/bash
set -e

REPO="andragon31/fenrir"
INSTALL_DIR="/usr/local/bin"

if [[ "$OSTYPE" == "darwin"* ]]; then
    ARCH=$(uname -m)
    if [ "$ARCH" = "arm64" ]; then
        BIN="fenrir-darwin-arm64"
    else
        BIN="fenrir-darwin-amd64"
    fi
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    ARCH=$(uname -m)
    if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        BIN="fenrir-linux-arm64"
    else
        BIN="fenrir-linux-amd64"
    fi
else
    echo "Unsupported OS: $OSTYPE"
    exit 1
fi

TMP=$(mktemp)
URL="https://github.com/${REPO}/releases/latest/download/${BIN}"

echo "Downloading Fenrir..."
curl -fsSL "$URL" -o "$TMP"
chmod +x "$TMP"

if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP" "$INSTALL_DIR/fenrir"
    echo "Installed to $INSTALL_DIR/fenrir"
else
    echo "Installing to $INSTALL_DIR requires sudo..."
    sudo mv "$TMP" "$INSTALL_DIR/fenrir"
    echo "Installed to $INSTALL_DIR/fenrir"
fi

echo ""
echo "Fenrir installed! Run:"
echo "  fenrir version          # Verify"
echo "  fenrir setup [agent]   # Setup for your AI tool"
echo ""
