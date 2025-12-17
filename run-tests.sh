#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Running Go Tests ===${NC}\n"

# Test Auth Service
echo -e "${YELLOW}Testing Auth Service...${NC}"
cd auth-service
go test -v -cover
AUTH_EXIT=$?
cd ..

echo ""

# Test Book Service
echo -e "${YELLOW}Testing Book Service...${NC}"
cd book-service
go test -v -cover
BOOK_EXIT=$?
cd ..

echo ""

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
if [ $AUTH_EXIT -eq 0 ]; then
    echo -e "${GREEN}✓ Auth Service tests passed${NC}"
else
    echo -e "${RED}✗ Auth Service tests failed${NC}"
fi

if [ $BOOK_EXIT -eq 0 ]; then
    echo -e "${GREEN}✓ Book Service tests passed${NC}"
else
    echo -e "${RED}✗ Book Service tests failed${NC}"
fi

echo ""

# Exit with error if any test failed
if [ $AUTH_EXIT -ne 0 ] || [ $BOOK_EXIT -ne 0 ]; then
    exit 1
fi

echo -e "${GREEN}All tests passed!${NC}"
