#!/bin/bash
set -e

# RevieU Secrets Sealing Script
# Encrypts secrets.yaml using Sealed Secrets public key from infra repo

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# GitHub raw URL for public key
INFRA_REPO_URL="https://raw.githubusercontent.com/RevieU-Corp/revieu-infra/main/k8s/pub-cert.pem"
LOCAL_PUB_KEY="/tmp/revieu-pub-cert.pem"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🔐 RevieU Secrets Sealing Tool"
echo "================================"

# Check if kubeseal is installed
if ! command -v kubeseal &> /dev/null; then
    echo -e "${RED}Error: kubeseal is not installed${NC}"
    echo ""
    echo "Install kubeseal:"
    echo "  macOS:   brew install kubeseal"
    echo "  Linux:   wget https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.5/kubeseal-0.24.5-linux-amd64.tar.gz"
    echo "           tar xfz kubeseal-0.24.5-linux-amd64.tar.gz && sudo install -m 755 kubeseal /usr/local/bin/kubeseal"
    exit 1
fi

# Download public key
echo "📥 Downloading public key from infra repo..."
if curl -fsSL "$INFRA_REPO_URL" -o "$LOCAL_PUB_KEY"; then
    echo -e "${GREEN}✅ Public key downloaded${NC}"
else
    echo -e "${RED}❌ Failed to download public key${NC}"
    exit 1
fi

# Service to seal (default: core)
SERVICE="${1:-core}"
SECRETS_FILE="$PROJECT_ROOT/apps/$SERVICE/configs/secrets.yaml"
OUTPUT_FILE="$PROJECT_ROOT/apps/$SERVICE/configs/sealed-secrets.yaml"

# Check if secrets file exists
if [ ! -f "$SECRETS_FILE" ]; then
    echo -e "${RED}Error: Secrets file not found at $SECRETS_FILE${NC}"
    echo ""
    echo "Usage: $0 [service-name]"
    exit 1
fi

echo ""
echo "📄 Input:  $SECRETS_FILE"
echo "🔑 Key:    Downloaded from infra repo"
echo "📦 Output: $OUTPUT_FILE"
echo ""

# Seal the secrets
echo "🔒 Encrypting secrets..."
if kubeseal --cert "$LOCAL_PUB_KEY" --format yaml < "$SECRETS_FILE" > "$OUTPUT_FILE"; then
    echo -e "${GREEN}✅ Secrets sealed successfully!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Review: cat $OUTPUT_FILE"
    echo "  2. Commit: git add $OUTPUT_FILE && git commit -m 'chore(core): update sealed secrets'"
    echo "  3. Push: git push"
    rm -f "$LOCAL_PUB_KEY"
else
    echo -e "${RED}❌ Failed to seal secrets${NC}"
    rm -f "$LOCAL_PUB_KEY"
    exit 1
fi
