#!/bin/bash
set -e

echo "🚀 Setting up RevieU Backend development environment..."
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.24+ first."
    exit 1
fi

# Install lefthook
echo "📦 Installing lefthook..."
go install github.com/evilmartians/lefthook@latest

# Install git hooks
echo "🪝 Installing git hooks..."
lefthook install

echo ""
echo "✅ Setup complete!"
echo ""
echo "Git hooks installed:"
echo "  - pre-commit: Prevents committing unencrypted secrets"
echo "  - commit-msg: Enforces subject, body, Closes, and Co-Authored-By"
echo "  - pre-push: Runs tests before pushing"
echo ""
echo "You're ready to start developing! 🎉"
