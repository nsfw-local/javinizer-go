#!/usr/bin/env bash
# Config synchronization validator
# Ensures config.yaml.example matches DefaultConfig() values
#
# Usage: ./scripts/validate-config-sync.sh
# Exit codes: 0 = sync OK, 1 = drift detected

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Validating config synchronization...${NC}"
echo ""

# Check 1: Verify config files exist
if [[ ! -f "configs/config.yaml.example" ]]; then
    echo -e "${RED}✗ configs/config.yaml.example not found${NC}"
    exit 1
fi

if [[ ! -f "internal/config/defaults.go" ]]; then
    echo -e "${RED}✗ internal/config/defaults.go not found${NC}"
    exit 1
fi

# Check 2: Run the Go drift detection test
echo -e "${BLUE}[1/3] Running drift detection test...${NC}"
if ! go test -short -run TestDefaultConfigMatchesExample ./internal/config/... 2>&1; then
    echo -e "${RED}✗ Drift detection test failed${NC}"
    echo ""
    echo -e "${YELLOW}This means config.yaml.example and DefaultConfig() have diverged.${NC}"
    echo -e "${YELLOW}Run 'go test -v -run TestDefaultConfigMatchesExample ./internal/config/...' for details.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Drift detection test passed${NC}"
echo ""

# Check 3: Verify config.yaml.example is valid YAML
echo -e "${BLUE}[2/3] Validating config.yaml.example YAML syntax...${NC}"
if command -v python3 &> /dev/null; then
    if ! python3 -c "import yaml; yaml.safe_load(open('configs/config.yaml.example'))" 2>&1; then
        echo -e "${RED}✗ config.yaml.example is not valid YAML${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ YAML syntax OK${NC}"
else
    echo -e "${YELLOW}⚠ Skipping YAML validation (python3 not available)${NC}"
fi
echo ""

# Check 4: Verify embedded config can be loaded
echo -e "${BLUE}[3/3] Validating embedded config matches source file...${NC}"
if ! go test -short -run TestEmbeddedConfigMatchesSourceFile ./internal/config/... 2>&1; then
    echo -e "${RED}✗ Embedded config validation failed${NC}"
    echo -e "${YELLOW}The embedded config may be out of sync with config.yaml.example${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Embedded config validation passed${NC}"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Config synchronization validation passed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}config.yaml.example and DefaultConfig() are synchronized.${NC}"

exit 0
