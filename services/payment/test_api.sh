#!/bin/bash

# Set base URL for the API Gateway
BASE_URL="http://localhost:8081/api/v1"

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "🚀 Starting Payment API tests..."

# Generate a test order ID
TEST_ORDER_ID=$(uuidgen)

echo -e "\n1. Testing Payment Creation"
PAYMENT_RESPONSE=$(curl -s -X POST "${BASE_URL}/payments" \
  -H "Content-Type: application/json" \
  -d "{
    \"order_id\": \"${TEST_ORDER_ID}\",
    \"amount\": 99.99,
    \"currency\": \"USD\",
    \"payment_method\": \"credit_card\"
  }")

echo "Response: ${PAYMENT_RESPONSE}"

# Extract payment ID from response
PAYMENT_ID=$(echo ${PAYMENT_RESPONSE} | jq -r '.ID')

if [ "${PAYMENT_ID}" != "null" ] && [ "${PAYMENT_ID}" != "" ]; then
    echo -e "${GREEN}✓ Payment creation successful${NC}"
else
    echo -e "${RED}✗ Payment creation failed${NC}"
    exit 1
fi

echo -e "\n2. Testing Get Payment"
GET_RESPONSE=$(curl -s -X GET "${BASE_URL}/payments/${PAYMENT_ID}")
echo "Response: ${GET_RESPONSE}"

RETRIEVED_ID=$(echo ${GET_RESPONSE} | jq -r '.ID')
if [ "${RETRIEVED_ID}" == "${PAYMENT_ID}" ]; then
    echo -e "${GREEN}✓ Payment retrieval successful${NC}"
else
    echo -e "${RED}✗ Payment retrieval failed${NC}"
    exit 1
fi

echo -e "\n3. Testing Invalid Payment ID"
INVALID_RESPONSE=$(curl -s -X GET "${BASE_URL}/payments/invalid-id")
echo "Response: ${INVALID_RESPONSE}"

ERROR_CODE=$(echo ${INVALID_RESPONSE} | jq -r '.code')
if [ "${ERROR_CODE}" == "INVALID_PAYMENT_ID" ]; then
    echo -e "${GREEN}✓ Invalid payment ID test passed${NC}"
else
    echo -e "${RED}✗ Invalid payment ID test failed${NC}"
    exit 1
fi

echo -e "\n✅ All payment API tests completed successfully!" 