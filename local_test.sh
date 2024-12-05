#!/bin/bash

API_URL="http://localhost:8080"
TOKEN="token-abc-123"

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

banner() {
  clear
  echo -e "                                                                          "
  echo -e "                 _____          ________   _______  ______ _____ _______  "
  echo -e "     ////////   / ____|        |  ____\ \ / /  __ \|  ____|  __ \__   __| "
  echo -e "               | |  __  ___    | |__   \ V /| |__) | |__  | |__) | | |    "
  echo -e " ////////////  | | |_ |/ _ \   |  __|   > < |  ___/|  __| |  _  /  | |    "
  echo -e "               | |__| | (_) |  | |____ / . \| |    | |____| | \ \  | |    "
  echo -e "        /////   \_____|\___/   |______/_/ \_\_|    |______|_|  \_\ |_|    "
  echo -e "                                                                          "
  echo -e "                                                                          "
}

test_endpoint() {
    local test_name=$1
    local expected_status=$2
    local use_token=$3
    local response
    local status_code

    if [ "$use_token" = true ]; then
        response=$(curl -s -o /dev/null -w "%{http_code}" -H "API_KEY: $TOKEN" $API_URL)
    else
        response=$(curl -s -o /dev/null -w "%{http_code}" $API_URL)
    fi

    status_code=$response
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✓ $test_name: Success${NC}"
    else
        echo -e "${RED}✗ $test_name: Expected $expected_status, got $status_code${NC}"
    fi
}

banner

echo "Starting Rate Limiter Tests..."

echo -e "\n1. Testing IP-based rate limiting"
echo "Making requests until blocked..."
for i in {1..6}; do
    test_endpoint "Request $i/5" "200" false
done
test_endpoint "Request after limit (should be blocked)" "429" false

echo -e "\n2. Testing Token-based rate limiting"
echo "Making requests until blocked..."
for i in {1..11}; do
    test_endpoint "Request $i/10" "200" true
done
test_endpoint "Request after limit (should be blocked)" "429" true

echo -e "\n3. Testing immediate block status"
test_endpoint "IP still blocked" "429" false
test_endpoint "Token still blocked" "429" true

echo -e "\nTests completed!"