#!/bin/bash

# Base URL for the API Gateway
BASE_URL="http://localhost:8081/api/v1/inventory"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🚀 Starting Inventory API tests..."

# Create a test product first to use its ID
TEST_PRODUCT_RESPONSE=$(curl -s -X POST "http://localhost:8081/api/v1/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "price": 99.99,
    "stock": 100,
    "images": ["test1.jpg", "test2.jpg"]
  }')

# Try both uppercase and lowercase ID fields
PRODUCT_ID=$(echo $TEST_PRODUCT_RESPONSE | jq -r '.id // .ID')
if [ -z "$PRODUCT_ID" ] || [ "$PRODUCT_ID" == "null" ]; then
    echo -e "${RED}Failed to create test product${NC}"
    exit 1
fi

echo -e "\n${GREEN}1. Testing Create Inventory Item${NC}"
CREATE_RESPONSE=$(curl -s -X POST "${BASE_URL}/items" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "'"${PRODUCT_ID}"'",
    "quantity": 100,
    "reorder_point": 20,
    "reorder_quantity": 50,
    "location": "WAREHOUSE-A1",
    "batch_number": "BATCH-001",
    "expiry_date": "2025-12-31",
    "notes": "Initial stock"
  }')
echo "Response: $CREATE_RESPONSE"

# Extract inventory item ID
INVENTORY_ID=$(echo "$CREATE_RESPONSE" | jq -r '.ID')
if [ -z "$INVENTORY_ID" ] || [ "$INVENTORY_ID" == "null" ] || [[ ! "$INVENTORY_ID" =~ ^[0-9a-fA-F]{8}- ]]; then
    echo -e "${RED}Failed to create inventory item - invalid ID: $INVENTORY_ID${NC}"
    exit 1
fi

echo -e "\n${GREEN}2. Testing Get Inventory Item${NC}"
GET_RESPONSE=$(curl -s -X GET "${BASE_URL}/items/${INVENTORY_ID}")
echo "Response: $GET_RESPONSE"

echo -e "\n${GREEN}3. Testing Update Stock (Receive)${NC}"
RECEIVE_RESPONSE=$(curl -s -X PUT "${BASE_URL}/items/${INVENTORY_ID}/stock" \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 50,
    "type": "received",
    "reference": "PO-001",
    "notes": "Received new stock"
  }')
echo "Response: $RECEIVE_RESPONSE"

echo -e "\n${GREEN}4. Testing Update Stock (Ship)${NC}"
SHIP_RESPONSE=$(curl -s -X PUT "${BASE_URL}/items/${INVENTORY_ID}/stock" \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 20,
    "type": "shipped",
    "reference": "ORDER-001",
    "notes": "Order fulfilled"
  }')
echo "Response: $SHIP_RESPONSE"

echo -e "\n${GREEN}5. Testing List All Inventory${NC}"
LIST_RESPONSE=$(curl -s -X GET "${BASE_URL}/items")
echo "$LIST_RESPONSE"

echo -e "\n${GREEN}6. Testing Get Transaction History${NC}"
HISTORY_RESPONSE=$(curl -s -X GET "${BASE_URL}/items/${INVENTORY_ID}/transactions")
echo "$HISTORY_RESPONSE"

echo -e "\n${GREEN}7. Testing Invalid Stock Update${NC}"
INVALID_RESPONSE=$(curl -s -X PUT "${BASE_URL}/items/${INVENTORY_ID}/stock" \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": -50,
    "type": "shipped",
    "reference": "INVALID-001",
    "notes": "This should fail"
  }')
if [[ "$INVALID_RESPONSE" == *"error"* ]]; then
    echo "Invalid stock update test passed!"
    echo "Response: $INVALID_RESPONSE"
else
    echo "Invalid stock update test failed!"
    echo "Response: $INVALID_RESPONSE"
fi

echo -e "\n✅ Inventory API tests completed\n"