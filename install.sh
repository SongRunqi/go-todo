#!/bin/bash

# Todo-Go Installation Script
# This script installs todo-go to your system

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default installation directory
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Todo-Go Installation Script       ║${NC}"
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go is not installed${NC}"
    echo -e "${YELLOW}  Please install Go from https://golang.org/dl/${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Go is installed: $(go version)${NC}"

# Build the application
echo ""
echo -e "${BLUE}Building todo-go...${NC}"
if go build -ldflags="-s -w" -o todo main.go; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# Create installation directory if it doesn't exist
if [ ! -d "$INSTALL_DIR" ]; then
    echo ""
    echo -e "${BLUE}Creating installation directory: $INSTALL_DIR${NC}"
    mkdir -p "$INSTALL_DIR"
fi

# Install the binary
echo ""
echo -e "${BLUE}Installing to $INSTALL_DIR/todo${NC}"
if mv todo "$INSTALL_DIR/todo"; then
    chmod +x "$INSTALL_DIR/todo"
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
if "$INSTALL_DIR/todo" init; then
    echo -e "${GREEN}✓ Initialization complete${NC}"
else
    echo -e "${YELLOW}⚠ Initialization skipped or failed${NC}"
fi

echo ""
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║    Installation Complete! 🎉           ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo -e "  1. Run ${GREEN}todo init${NC} to set up your configuration"
echo -e "  2. Run ${GREEN}todo lang set en${NC} or ${GREEN}todo lang set zh${NC} to set language"
echo -e "  3. Run ${GREEN}todo --help${NC} to see available commands"
echo -e "  4. Create your first task: ${GREEN}todo \"买菜 明天下午5点截止\"${NC}"
echo ""
