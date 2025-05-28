#!/bin/bash

# Script Writing Workflow Runner
# This script makes it easy to run the script writing workflow

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}🎬 Script Writing Workflow (1000 Token Limit)${NC}"
echo ""

# Check if ANTHROPIC_API_KEY is set
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo -e "${RED}❌ Error: ANTHROPIC_API_KEY environment variable is not set${NC}"
    echo ""
    echo "Please set your Anthropic API key:"
    echo "  export ANTHROPIC_API_KEY='your-api-key-here'"
    echo ""
    echo "Or create a .env file with:"
    echo "  echo 'ANTHROPIC_API_KEY=your-api-key-here' > .env"
    echo "  source .env"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Error: Go is not installed${NC}"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

echo -e "${GREEN}✓ Environment checks passed${NC}"
echo ""

# If no arguments provided, show usage
if [ $# -eq 0 ]; then
    echo -e "${YELLOW}Usage:${NC}"
    echo "  $0 \"Your script description here\""
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 \"A romantic comedy about two coffee shop employees\""
    echo "  $0 \"A thriller scene in an abandoned warehouse\""
    echo "  $0 \"A drama about a family reunion gone wrong\""
    echo ""
    echo -e "${YELLOW}Running with default example...${NC}"
    echo ""
    
    # Run with default
    go run main.go
else
    # Run with provided description
    echo -e "${YELLOW}Description:${NC} $*"
    echo ""
    go run main.go "$*"
fi

echo ""
echo -e "${GREEN}✅ Workflow completed!${NC}"
echo ""
echo -e "${YELLOW}Check the scripts/ directory for your generated script.${NC}"
