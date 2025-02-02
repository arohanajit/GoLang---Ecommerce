#!/bin/bash

# Base URL for the API Gateway
GATEWAY_URL="http://localhost:8081/api/v1"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Generate a unique email using timestamp
TIMESTAMP=$(date +%s)
TEST_EMAIL="test${TIMESTAMP}@example.com"

# Function to check if a service is running
check_service() {
    local port=$1
    local service_name=$2
    nc -z localhost $port
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $service_name is running on port $port${NC}"
        return 0
    else
        echo -e "${RED}✗ $service_name is not running on port $port${NC}"
        return 1
    fi
}

# Check if jq is installed (for JSON parsing)
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is required. Install with 'brew install jq' (macOS) or 'sudo apt-get install jq' (Linux).${NC}"
    exit 1
fi

echo -e "${BLUE}🔍 Checking if all services are running...${NC}"

# Check all services
check_service 8081 "API Gateway"
check_service 8002 "User Service" 
check_service 8000 "Product Service"  # Changed from 8003
check_service 8001 "Order Service"    # Changed from 8004  
check_service 8004 "Payment Service"  # Changed from 8005
check_service 8003 "Inventory Service" # Changed from 8006

echo -e "\n${GREEN}🚀 Starting comprehensive service tests...${NC}"

# First run the user service specific tests
echo -e "\n${BLUE}Running User Service specific tests...${NC}"
if [ -f "./services/user/test_api.sh" ]; then
    chmod +x ./services/user/test_api.sh
    ./services/user/test_api.sh
    if [ $? -ne 0 ]; then
        echo -e "${RED}User Service specific tests failed${NC}"
        exit 1
    fi
else
    echo -e "${RED}Warning: User Service test script not found at ./services/user/test_api.sh${NC}"
fi

# Run payment service specific tests
echo -e "\n${BLUE}Running Payment Service specific tests...${NC}"
if [ -f "./services/payment/test_api.sh" ]; then
    chmod +x ./services/payment/test_api.sh
    ./services/payment/test_api.sh
    if [ $? -ne 0 ]; then
        echo -e "${RED}Payment Service specific tests failed${NC}"
        exit 1
    fi
else
    echo -e "${RED}Warning: Payment Service test script not found at ./services/payment/test_api.sh${NC}"
fi

# Test User Registration
echo -e "\n${GREEN}1. Testing User Registration${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/users/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "'"${TEST_EMAIL}"'",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "phone_number": "+1234567890"
  }')

# Check if response contains error
if echo "$REGISTER_RESPONSE" | grep -q '"error"'; then
    ERROR_MESSAGE=$(echo "$REGISTER_RESPONSE" | jq -r '.error')
    echo -e "${RED}Error: $ERROR_MESSAGE${NC}"
    exit 1
fi
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.user_id')
# Check if registration was successful
if echo "$REGISTER_RESPONSE" | grep -q '"message\|user_id"'; then
    echo -e "${GREEN}User registered successfully!${NC}"
    USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.user_id')
else
    echo -e "${RED}Error: User registration failed.${NC}"
    echo "Full response: $REGISTER_RESPONSE"
    exit 1
fi

# Test Login
echo -e "\n${GREEN}2. Testing Login${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/users/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "'"${TEST_EMAIL}"'",
    "password": "password123"
  }')
echo "Response: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

# Verify we got a valid token
if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}Error: Failed to get valid token${NC}"
    exit 1
fi

# Test Product Creation
echo -e "\n${GREEN}3. Testing Product Creation${NC}"
PRODUCT_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/products" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "MacBook Pro",
    "price": 1999.99,
    "stock": 10,
    "images": ["image1.jpg", "image2.jpg"]
  }')
echo "Response: $PRODUCT_RESPONSE"
PRODUCT_ID=$(echo "$PRODUCT_RESPONSE" | jq -r '.id')

# Check if product ID is in UUID format
if [[ ! "$PRODUCT_ID" =~ ^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$ ]]; then
    echo -e "${RED}Invalid product ID format: $PRODUCT_ID${NC}"
    exit 1
fi

# Test Order Creation
echo -e "\n${GREEN}4. Testing Order Creation${NC}"
ORDER_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "user_id": "'"$USER_ID"'",
    "items": [
      {
        "product_id": "'"$PRODUCT_ID"'",
        "quantity": 2
      }
    ]
  }')
echo "Response: $ORDER_RESPONSE"
ORDER_ID=$(echo $ORDER_RESPONSE | jq -r '.ID')

# Test Payment Creation for Order
echo -e "\n${GREEN}5. Testing Payment Creation${NC}"
PAYMENT_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/payments" \
  -H "Content-Type: application/json" \
  -d "{
    \"order_id\": \"${ORDER_ID}\",
    \"amount\": 1999.99,
    \"currency\": \"USD\",
    \"payment_method\": \"credit_card\"
  }")
echo "Response: $PAYMENT_RESPONSE"

# Extract payment ID and verify
PAYMENT_ID=$(echo ${PAYMENT_RESPONSE} | jq -r '.ID')
if [ "${PAYMENT_ID}" != "null" ] && [ "${PAYMENT_ID}" != "" ]; then
    echo -e "${GREEN}✓ Payment creation successful${NC}"
else
    echo -e "${RED}✗ Payment creation failed${NC}"
    exit 1
fi

# Test Payment Retrieval
echo -e "\n${GREEN}6. Testing Payment Retrieval${NC}"
GET_PAYMENT_RESPONSE=$(curl -s -X GET "${GATEWAY_URL}/payments/${PAYMENT_ID}")
echo "Response: $GET_PAYMENT_RESPONSE"

RETRIEVED_PAYMENT_ID=$(echo ${GET_PAYMENT_RESPONSE} | jq -r '.ID')
if [ "${RETRIEVED_PAYMENT_ID}" == "${PAYMENT_ID}" ]; then
    echo -e "${GREEN}✓ Payment retrieval successful${NC}"
else
    echo -e "${RED}✗ Payment retrieval failed${NC}"
    exit 1
fi

# Test Invalid Payment ID
echo -e "\n${GREEN}7. Testing Invalid Payment ID${NC}"
INVALID_PAYMENT_RESPONSE=$(curl -s -X GET "${GATEWAY_URL}/payments/invalid-id")
echo "Response: $INVALID_PAYMENT_RESPONSE"

PAYMENT_ERROR_CODE=$(echo ${INVALID_PAYMENT_RESPONSE} | jq -r '.code')
if [ "${PAYMENT_ERROR_CODE}" == "INVALID_PAYMENT_ID" ]; then
    echo -e "${GREEN}✓ Invalid payment ID test passed${NC}"
else
    echo -e "${RED}✗ Invalid payment ID test failed${NC}"
    exit 1
fi

# Test Get Order with correct ORDER_ID
echo -e "\n${GREEN}8. Testing Get Order${NC}"
curl -s -X GET "${GATEWAY_URL}/orders/${ORDER_ID}" \
  -H "Authorization: Bearer $TOKEN"

# Test Get Product
echo -e "\n${GREEN}9. Testing Get Product${NC}"
curl -s -X GET "${GATEWAY_URL}/products/${PRODUCT_ID}"

# Test Invalid Order Fetch
echo -e "\n${GREEN}10. Testing Invalid Order Fetch${NC}"
INVALID_ORDER_RESPONSE=$(curl -s -X GET "${GATEWAY_URL}/orders/invalid_order_id" \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $INVALID_ORDER_RESPONSE"

# Check if error response matches the expected format
ERROR_CODE=$(echo "$INVALID_ORDER_RESPONSE" | jq -r '.code')
if [ "$ERROR_CODE" = "INVALID_ORDER_ID" ]; then
    echo -e "${GREEN}Invalid order ID test passed!${NC}"
else
    echo -e "${RED}Invalid order ID test failed!${NC}"
    echo "Expected error code: INVALID_ORDER_ID"
    echo "Actual response: $INVALID_ORDER_RESPONSE"
    exit 1
fi

# Test Invalid Product ID
echo -e "\n${GREEN}11. Testing Invalid Product ID${NC}"
INVALID_PRODUCT_RESPONSE=$(curl -s -X GET "${GATEWAY_URL}/products/invalid_product_id")
echo "Response: $INVALID_PRODUCT_RESPONSE"

# Check if error response matches the expected format
ERROR_CODE=$(echo "$INVALID_PRODUCT_RESPONSE" | jq -r '.code')
if [ "$ERROR_CODE" = "INVALID_PRODUCT_ID" ]; then
    echo -e "${GREEN}Invalid product ID test passed!${NC}"
else
    echo -e "${RED}Invalid product ID test failed!${NC}"
    echo "Expected error code: INVALID_PRODUCT_ID"
    echo "Actual response: $INVALID_PRODUCT_RESPONSE"
fi

# Test User Profile
echo -e "\n${GREEN}12. Testing Get User Profile${NC}"
PROFILE_RESPONSE=$(curl -s -X GET "${GATEWAY_URL}/users/profile" \
  -H "Authorization: Bearer $TOKEN")
echo "Profile Response: $PROFILE_RESPONSE"

# Test Address Creation
echo -e "\n${GREEN}13. Testing Address Creation${NC}"
ADDRESS_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/users/addresses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "street": "123 Test St",
    "city": "Test City",
    "state": "TS",
    "country": "Test Country",
    "postal_code": "12345",
    "is_default": true
  }')
echo "Address Response: $ADDRESS_RESPONSE"

echo -e "\n${GREEN}✅ All tests completed successfully!${NC}"