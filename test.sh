#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Traefik Forward Auth POC - Test Script ===${NC}\n"

# Wait for services to be ready
echo -e "${YELLOW}Waiting for services to be ready...${NC}"
sleep 3

BASE_URL="http://localhost"

echo -e "\n${BLUE}Test 1: Request without authentication${NC}"
echo "Expected: 401 Unauthorized"
curl -s -o /dev/null -w "Status: %{http_code}\n" $BASE_URL/books
echo ""

echo -e "${BLUE}Test 2: Request with invalid token${NC}"
echo "Expected: 401 Unauthorized"
curl -s -o /dev/null -w "Status: %{http_code}\n" \
  -H "Authorization: Bearer user:admin:wrong-token" \
  $BASE_URL/books
echo ""

echo -e "${BLUE}Test 3: Request with Admin token${NC}"
echo "Expected: 200 OK with all 6 books"
response=$(curl -s -H "Authorization: Bearer admin:admin:secret-token" $BASE_URL/books)
status=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer admin:admin:secret-token" $BASE_URL/books)
total_books=$(echo $response | jq -r '.total_books')
echo "Status: $status"
echo "Total books: $total_books"
echo "Response:"
echo $response | jq '.'
echo ""

echo -e "${BLUE}Test 4: Request with User token${NC}"
echo "Expected: 200 OK with 4 books (public + user)"
response=$(curl -s -H "Authorization: Bearer alice:user:secret-token" $BASE_URL/books)
status=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer alice:user:secret-token" $BASE_URL/books)
total_books=$(echo $response | jq -r '.total_books')
echo "Status: $status"
echo "Total books: $total_books"
echo "Response:"
echo $response | jq '.'
echo ""

echo -e "${BLUE}Test 5: Request with Guest token${NC}"
echo "Expected: 200 OK with 2 books (public only)"
response=$(curl -s -H "Authorization: Bearer bob:guest:secret-token" $BASE_URL/books)
status=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer bob:guest:secret-token" $BASE_URL/books)
total_books=$(echo $response | jq -r '.total_books')
echo "Status: $status"
echo "Total books: $total_books"
echo "Response:"
echo $response | jq '.'
echo ""

echo -e "${BLUE}Test 6: View access logs${NC}"
echo "Expected: List of all authentication attempts"
curl -s $BASE_URL/access-logs | jq '.'
echo ""

echo -e "${GREEN}=== All tests completed ===${NC}"
echo -e "\n${YELLOW}Additional information:${NC}"
echo "- Traefik Dashboard: http://localhost:8080"
echo "- Auth Service (direct): http://localhost:8081"
echo "- Book Service (direct): http://localhost:8082"
