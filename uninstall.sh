#!/bin/bash

# Todo-Go Uninstallation Script
# This script uninstalls todo-go from your system

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default installation directory
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
TODO_DIR="$HOME/.todo"

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Todo-Go Uninstallation Script     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# Check if todo binary exists
if [ ! -f "$INSTALL_DIR/todo" ]; then
    echo -e "${YELLOW}⚠ Todo binary not found at $INSTALL_DIR/todo${NC}"
    echo -e "${YELLOW}  It may have already been uninstalled or installed in a different location.${NC}"
    echo ""

    # Ask if user wants to remove data anyway
    if [ -d "$TODO_DIR" ]; then
        read -p "$(echo -e ${BLUE}Remove todo data directory \($TODO_DIR\)? \[y/N\]: ${NC})" -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$TODO_DIR"
            echo -e "${GREEN}✓ Todo data directory removed${NC}"
        else
            echo -e "${YELLOW}⚠ Todo data directory kept${NC}"
        fi
    fi
    exit 0
fi

echo -e "${BLUE}Found todo installation at: $INSTALL_DIR/todo${NC}"
echo ""

# Ask for confirmation
read -p "$(echo -e ${YELLOW}Are you sure you want to uninstall todo-go? \[y/N\]: ${NC})" -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}Uninstallation cancelled.${NC}"
    exit 0
fi

echo ""

# Remove the binary
echo -e "${BLUE}Removing todo binary...${NC}"
if rm "$INSTALL_DIR/todo"; then
    echo -e "${GREEN}✓ Binary removed from $INSTALL_DIR/todo${NC}"
else
    echo -e "${RED}✗ Failed to remove binary${NC}"
    exit 1
fi

# Ask about removing data directory
echo ""
if [ -d "$TODO_DIR" ]; then
    echo -e "${YELLOW}Todo data directory found at: $TODO_DIR${NC}"
    echo -e "${YELLOW}This contains your tasks, completed tasks, and configuration.${NC}"
    echo ""
    read -p "$(echo -e ${BLUE}Do you want to remove the todo data directory? \[y/N\]: ${NC})" -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Backup before removing
        BACKUP_NAME="todo_backup_$(date +%Y%m%d_%H%M%S).tar.gz"
        echo ""
        echo -e "${BLUE}Creating backup at ~/$BACKUP_NAME...${NC}"
        if tar -czf "$HOME/$BACKUP_NAME" -C "$HOME" .todo 2>/dev/null; then
            echo -e "${GREEN}✓ Backup created: ~/$BACKUP_NAME${NC}"
            echo ""
            rm -rf "$TODO_DIR"
            echo -e "${GREEN}✓ Todo data directory removed${NC}"
        else
            echo -e "${YELLOW}⚠ Backup failed, but continuing with removal...${NC}"
            rm -rf "$TODO_DIR"
            echo -e "${GREEN}✓ Todo data directory removed${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ Todo data directory kept at $TODO_DIR${NC}"
        echo -e "${YELLOW}  You can manually remove it later if needed.${NC}"
    fi
else
    echo -e "${BLUE}ℹ No todo data directory found${NC}"
fi

# Check for PATH modifications
echo ""
echo -e "${BLUE}Checking shell configuration files...${NC}"
SHELL_CONFIGS=("$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.zshrc" "$HOME/.config/fish/config.fish")
PATH_FOUND=false

for config in "${SHELL_CONFIGS[@]}"; do
    if [ -f "$config" ] && grep -q "$INSTALL_DIR" "$config"; then
        echo -e "${YELLOW}⚠ Found reference to $INSTALL_DIR in $config${NC}"
        PATH_FOUND=true
    fi
done

if [ "$PATH_FOUND" = true ]; then
    echo ""
    echo -e "${YELLOW}Note: You may want to remove the PATH export from your shell configuration.${NC}"
    echo -e "${YELLOW}Look for lines like: export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
fi

echo ""
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║    Uninstallation Complete! 👋        ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Thank you for using Todo-Go!${NC}"
echo ""
