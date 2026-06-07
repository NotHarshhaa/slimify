#!/bin/sh
# install.sh — install slimify from GitHub releases
#
# Usage:
#   curl -sSfL https://raw.githubusercontent.com/NotHarshhaa/slimify/main/install.sh | sh
#   curl -sSfL https://raw.githubusercontent.com/NotHarshhaa/slimify/main/install.sh | sh -s -- -b /usr/local/bin
#   curl -sSfL https://raw.githubusercontent.com/NotHarshhaa/slimify/main/install.sh | sh -s -- -v v1.0.0

set -e

REPO="NotHarshhaa/slimify"
BINARY="slimify"
INSTALL_DIR="/usr/local/bin"
VERSION=""

# Parse flags
while [ $# -gt 0 ]; do
    case "$1" in
        -b|--bin)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    linux)   OS="Linux" ;;
    darwin)  OS="Darwin" ;;
    mingw*|msys*|cygwin*) OS="Windows" ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Get latest version if not specified
if [ -z "$VERSION" ]; then
    VERSION=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        echo "Failed to fetch latest version"
        exit 1
    fi
fi

# Remove leading 'v' for filename
VERSION_NUM=$(echo "$VERSION" | sed 's/^v//')

# Determine file extension
EXT="tar.gz"
if [ "$OS" = "Windows" ]; then
    EXT="zip"
fi

# Build download URL
FILENAME="${BINARY}_${VERSION_NUM}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

echo "Installing ${BINARY} ${VERSION} (${OS}/${ARCH})..."
echo "Downloading: ${URL}"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

# Download
curl -sSfL "$URL" -o "${TMP_DIR}/${FILENAME}"

# Extract
cd "$TMP_DIR"
if [ "$EXT" = "zip" ]; then
    unzip -q "$FILENAME"
else
    tar xzf "$FILENAME"
fi

# Install
mkdir -p "$INSTALL_DIR"
if [ -w "$INSTALL_DIR" ]; then
    cp "${BINARY}" "$INSTALL_DIR/"
else
    sudo cp "${BINARY}" "$INSTALL_DIR/"
fi

chmod +x "$INSTALL_DIR/$BINARY"

echo ""
echo "✓ ${BINARY} ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Run '${BINARY} --help' to get started."
