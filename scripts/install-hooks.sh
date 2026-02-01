#!/bin/bash

# Install Git hooks for RevieU Backend

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🔧 Installing Git hooks..."
echo ""

# Check if .git directory exists
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo "Error: Not a git repository"
    exit 1
fi

# Install pre-commit hook
if [ -f "$HOOKS_DIR/pre-commit" ]; then
    echo -e "${YELLOW}⚠️  Pre-commit hook already exists${NC}"
    read -p "Overwrite? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Skipping pre-commit hook installation"
        exit 0
    fi
fi

# Copy and make executable
cp "$SCRIPT_DIR/hooks/pre-commit" "$HOOKS_DIR/pre-commit"
chmod +x "$HOOKS_DIR/pre-commit"

echo -e "${GREEN}✅ Pre-commit hook installed successfully!${NC}"
echo ""
echo "This hook will prevent you from committing unencrypted secrets.yaml files."
echo ""
echo "To test it:"
echo "  1. Try to commit a secrets.yaml file (it should be blocked)"
echo "  2. Use ./scripts/seal-secrets.sh to encrypt secrets"
echo "  3. Commit the sealed-secrets.yaml file instead"
