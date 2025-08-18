#!/bin/bash

# Test coverage script for Debian Doctor Go
# This script runs all tests and generates comprehensive coverage reports

set -euo pipefail

echo "🧪 Running comprehensive test suite with coverage..."

# Create coverage directory
mkdir -p coverage

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}📊 Running unit tests with coverage...${NC}"

# Run tests with coverage
go test -v -coverprofile=coverage/coverage.out ./...

# Check if coverage file was created
if [ ! -f coverage/coverage.out ]; then
    echo -e "${RED}❌ Failed to generate coverage report${NC}"
    exit 1
fi

# Generate HTML coverage report
echo -e "${BLUE}📈 Generating HTML coverage report...${NC}"
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# Generate coverage summary
echo -e "${BLUE}📋 Coverage Summary:${NC}"
go tool cover -func=coverage/coverage.out

# Extract total coverage percentage
TOTAL_COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo ""
echo -e "${BLUE}📊 Coverage Report Summary:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if (( $(echo "$TOTAL_COVERAGE >= 80" | bc -l) )); then
    echo -e "${GREEN}✅ Total Coverage: ${TOTAL_COVERAGE}% - EXCELLENT${NC}"
elif (( $(echo "$TOTAL_COVERAGE >= 70" | bc -l) )); then
    echo -e "${YELLOW}⚠️  Total Coverage: ${TOTAL_COVERAGE}% - GOOD${NC}"
elif (( $(echo "$TOTAL_COVERAGE >= 50" | bc -l) )); then
    echo -e "${YELLOW}⚠️  Total Coverage: ${TOTAL_COVERAGE}% - NEEDS IMPROVEMENT${NC}"
else
    echo -e "${RED}❌ Total Coverage: ${TOTAL_COVERAGE}% - POOR${NC}"
fi

echo ""
echo -e "${BLUE}📁 Coverage files generated:${NC}"
echo "  - Text report: coverage/coverage.out"
echo "  - HTML report: coverage/coverage.html"
echo ""

# Package-specific coverage
echo -e "${BLUE}📦 Package Coverage Breakdown:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

go tool cover -func=coverage/coverage.out | grep -E "^github.com/sonyccd/debian-doctor-go/" | \
while read line; do
    PACKAGE=$(echo "$line" | awk -F'/' '{print $NF}' | awk '{print $1}')
    COVERAGE=$(echo "$line" | awk '{print $3}')
    COVERAGE_NUM=$(echo "$COVERAGE" | sed 's/%//')
    
    if (( $(echo "$COVERAGE_NUM >= 80" | bc -l) )); then
        echo -e "  ${GREEN}✅ $PACKAGE: $COVERAGE${NC}"
    elif (( $(echo "$COVERAGE_NUM >= 60" | bc -l) )); then
        echo -e "  ${YELLOW}⚠️  $PACKAGE: $COVERAGE${NC}"
    else
        echo -e "  ${RED}❌ $PACKAGE: $COVERAGE${NC}"
    fi
done

echo ""

# Run benchmarks
echo -e "${BLUE}🏃 Running benchmarks...${NC}"
go test -bench=. -benchmem ./... > coverage/benchmarks.txt

echo -e "${GREEN}✅ All tests and coverage reports completed!${NC}"
echo ""
echo -e "${BLUE}🔍 To view detailed coverage:${NC}"
echo "  - Open coverage/coverage.html in your browser"
echo "  - View benchmarks: cat coverage/benchmarks.txt"
echo ""

# Check if we're in CI environment
if [ "${CI:-false}" = "true" ]; then
    echo -e "${BLUE}🤖 CI Environment detected - generating additional reports...${NC}"
    
    # Generate coverage badge info
    echo "$TOTAL_COVERAGE" > coverage/coverage_percentage.txt
    
    # Generate coverage for each package in JSON format for CI
    go tool cover -func=coverage/coverage.out | grep -E "^github.com/sonyccd/debian-doctor-go/" > coverage/package_coverage.txt
fi

echo -e "${GREEN}🎉 Test coverage analysis complete!${NC}"