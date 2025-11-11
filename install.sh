#!/bin/bash

# Todo-Go Installation Script
# Downloads and installs the latest release from GitHub

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="SongRunqi/go-todo"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="todo"

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Todo-Go Installation Script       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "darwin";;
        MINGW*|MSYS*|CYGWIN*)     echo "windows";;
        *)          echo "unknown";;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64";;
        aarch64|arm64)  echo "arm64";;
        *)              echo "unknown";;
    esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

echo -e "${BLUE}Detected platform:${NC} $OS-$ARCH"

# Validate platform
if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
    echo -e "${RED}✗ Unsupported platform: $OS-$ARCH${NC}"
    exit 1
fi

if [ "$OS" = "windows" ] && [ "$ARCH" = "arm64" ]; then
    echo -e "${RED}✗ Windows ARM64 is not currently supported${NC}"
    exit 1
fi

# Get latest release version
echo ""
echo -e "${BLUE}Fetching latest release information...${NC}"

# Try to get latest release from GitHub API
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest")

if echo "$LATEST_RELEASE" | grep -q '"message": "Not Found"'; then
    echo -e "${RED}✗ No releases found${NC}"
    echo -e "${YELLOW}  Please check https://github.com/$REPO/releases${NC}"
    exit 1
fi

VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo -e "${RED}✗ Failed to get version information${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Latest version: $VERSION${NC}"

# Construct download URLs
if [ "$OS" = "windows" ]; then
    BINARY_FILE="todo-${OS}-${ARCH}-todo.exe"
else
    BINARY_FILE="todo-${OS}-${ARCH}-todo"
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY_FILE"
CHECKSUM_URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY_FILE}.sha256"

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo ""
echo -e "${BLUE}Downloading $BINARY_FILE...${NC}"
if ! curl -L -o "$TMP_DIR/$BINARY_NAME" "$DOWNLOAD_URL" 2>/dev/null; then
    echo -e "${RED}✗ Download failed${NC}"
    echo -e "${YELLOW}  URL: $DOWNLOAD_URL${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Download complete${NC}"

# Download and verify checksum
echo ""
echo -e "${BLUE}Downloading checksum...${NC}"
if curl -L -s -o "$TMP_DIR/${BINARY_NAME}.sha256" "$CHECKSUM_URL" 2>/dev/null; then
    echo -e "${GREEN}✓ Checksum downloaded${NC}"

    echo -e "${BLUE}Verifying checksum...${NC}"
    EXPECTED_CHECKSUM=$(cat "$TMP_DIR/${BINARY_NAME}.sha256")

    if command -v sha256sum &> /dev/null; then
        ACTUAL_CHECKSUM=$(sha256sum "$TMP_DIR/$BINARY_NAME" | cut -d ' ' -f 1)
    elif command -v shasum &> /dev/null; then
        ACTUAL_CHECKSUM=$(shasum -a 256 "$TMP_DIR/$BINARY_NAME" | cut -d ' ' -f 1)
    else
        echo -e "${YELLOW}⚠ Warning: sha256sum not found, skipping checksum verification${NC}"
        ACTUAL_CHECKSUM=""
    fi

    if [ -n "$ACTUAL_CHECKSUM" ]; then
        if [ "$EXPECTED_CHECKSUM" = "$ACTUAL_CHECKSUM" ]; then
            echo -e "${GREEN}✓ Checksum verified${NC}"
        else
            echo -e "${RED}✗ Checksum verification failed${NC}"
            echo -e "${YELLOW}  Expected: $EXPECTED_CHECKSUM${NC}"
            echo -e "${YELLOW}  Got:      $ACTUAL_CHECKSUM${NC}"
            exit 1
        fi
    fi
else
    echo -e "${YELLOW}⚠ Warning: Could not download checksum, skipping verification${NC}"
fi

# Create installation directory if it doesn't exist
if [ ! -d "$INSTALL_DIR" ]; then
    echo ""
    echo -e "${BLUE}Creating installation directory: $INSTALL_DIR${NC}"
    mkdir -p "$INSTALL_DIR"
fi

# Install the binary
echo ""
echo -e "${BLUE}Installing to $INSTALL_DIR/$BINARY_NAME${NC}"
if mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"; then
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    echo -e "${GREEN}✓ Installation successful${NC}"
else
    echo -e "${RED}✗ Installation failed${NC}"
    exit 1
fi

# Check if install directory is in PATH
echo ""
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}⚠ Warning: $INSTALL_DIR is not in your PATH${NC}"
    echo ""
    echo -e "${YELLOW}Add the following line to your shell configuration file:${NC}"
    echo -e "${BLUE}  export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
    echo ""
    echo -e "${YELLOW}Shell configuration files:${NC}"
    echo -e "  ${BLUE}Bash:${NC}   ~/.bashrc or ~/.bash_profile"
    echo -e "  ${BLUE}Zsh:${NC}    ~/.zshrc"
    echo -e "  ${BLUE}Fish:${NC}   ~/.config/fish/config.fish"
    echo ""
fi

# Initialize todo directories
echo -e "${BLUE}Initializing todo directories...${NC}"
if "$INSTALL_DIR/$BINARY_NAME" init; then
    echo -e "${GREEN}✓ Initialization complete${NC}"
else
    echo -e "${YELLOW}⚠ Initialization skipped or failed${NC}"
fi

echo ""
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║    Installation Complete! 🎉           ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Installed version:${NC} $VERSION"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo -e "  1. Set your API key: ${GREEN}export API_KEY=\"your-deepseek-api-key\"${NC}"
echo -e "  2. Set language: ${GREEN}todo lang set en${NC} or ${GREEN}todo lang set zh${NC}"
echo -e "  3. View help: ${GREEN}todo --help${NC}"
echo -e "  4. Create your first task: ${GREEN}todo \"Buy groceries tomorrow\"${NC}"
echo ""
