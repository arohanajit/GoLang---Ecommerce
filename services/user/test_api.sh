#!/bin/bash

# Use gateway URL and enable error handling
BASE_URL="http://localhost:8081/api/v1/users"
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "🚀 Starting API tests..."

# Generate a unique test email using timestamp
TIMESTAMP=$(date +%s)
TEST_EMAIL="test${TIMESTAMP}@example.com"

# 1. Register a new user
echo -e "\n${GREEN}1. Testing User Registration${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "'"${TEST_EMAIL}"'",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "phone_number": "+1234567890"
  }')
echo "Response: $REGISTER_RESPONSE"

# Check if registration was successful
if ! echo "$REGISTER_RESPONSE" | grep -q "User registered successfully"; then
    echo -e "${RED}Registration failed${NC}"
    exit 1
fi

# Extract user_id from registration response
USER_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"user_id":"[^"]*' | grep -o '[^"]*$')

# 2. Login with the newly created user
echo -e "\n${GREEN}2. Testing Login${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "'"${TEST_EMAIL}"'",
    "password": "password123"
  }')
echo "Response: $LOGIN_RESPONSE"

# Extract token from login response
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | grep -o '[^"]*$')
if [ -z "$TOKEN" ]; then
    echo -e "${RED}Login failed - no token received${NC}"
    echo "Login response: $LOGIN_RESPONSE"
    exit 1
fi
echo "Token received successfully"

# 3. Get Profile
echo -e "\n${GREEN}3. Testing Get Profile${NC}"
curl -s -X GET "${BASE_URL}/profile" \
  -H "Authorization: Bearer $TOKEN"

# 4. Update Profile
echo -e "\n${GREEN}4. Testing Update Profile${NC}"
curl -s -X PUT "${BASE_URL}/profile" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John Updated",
    "last_name": "Doe Updated",
    "phone_number": "+1987654321",
    "date_of_birth": "1990-01-01T00:00:00Z",
    "profile_picture": "https://example.com/profile.jpg",
    "bio": "A software engineer who loves coding",
    "preferred_language": "en"
  }'

# 5. Add Address
echo -e "\n${GREEN}5. Testing Add Address${NC}"
ADDRESS_RESPONSE=$(curl -s -X POST "${BASE_URL}/addresses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "street": "123 Main St",
    "city": "New York",
    "state": "NY",
    "country": "USA",
    "postal_code": "10001"
  }')
echo "Response: $ADDRESS_RESPONSE"

# Extract address ID from response
ADDRESS_ID=$(echo "$ADDRESS_RESPONSE" | grep -o '"ID":[0-9]*' | grep -o '[0-9]*')

# 6. List Addresses
echo -e "\n${GREEN}6. Testing List Addresses${NC}"
curl -s -X GET "${BASE_URL}/addresses" \
  -H "Authorization: Bearer $TOKEN"

# 7. Update Address
echo -e "\n${GREEN}7. Testing Update Address${NC}"
curl -s -X PUT "${BASE_URL}/addresses/${ADDRESS_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "street": "456 Updated St",
    "city": "Updated City",
    "state": "UC",
    "country": "USA",
    "postal_code": "20002"
  }'

# 8. Change Password
echo -e "\n${GREEN}8. Testing Change Password${NC}"
curl -s -X PUT "${BASE_URL}/profile/change-password" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "password123",
    "new_password": "newpassword123"
  }'

# 9. Delete Address
echo -e "\n${GREEN}9. Testing Delete Address${NC}"
curl -s -X DELETE "${BASE_URL}/addresses/${ADDRESS_ID}" \
  -H "Authorization: Bearer $TOKEN"

# 10. Delete Account
echo -e "\n${GREEN}10. Testing Delete Account${NC}"
curl -s -X DELETE "${BASE_URL}/profile" \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n${GREEN}✅ API tests completed${NC}"
