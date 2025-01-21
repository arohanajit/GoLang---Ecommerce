#!/bin/bash

# Base URL
BASE_URL="http://localhost:8080"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "🚀 Starting API tests..."

# 1. Register a new user
echo -e "\n${GREEN}1. Testing User Registration${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "phone_number": "+1234567890"
  }')
echo "Response: $REGISTER_RESPONSE"

# 2. Login
echo -e "\n${GREEN}2. Testing Login${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')
echo "Response: $LOGIN_RESPONSE"

# Extract token from login response
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | grep -o '[^"]*$')
echo "Token: $TOKEN"

# 3. Request Password Reset
echo -e "\n${GREEN}3. Testing Password Reset Request${NC}"
curl -s -X POST "${BASE_URL}/forgot-password" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'

# Note: In a real test environment, you would need to extract the reset token from the email
# For testing purposes, you can temporarily modify the code to return the token in the response
# or query it directly from the database

# 4. Reset Password (replace RESET_TOKEN with actual token)
echo -e "\n${GREEN}4. Testing Password Reset${NC}"
echo "Note: Replace RESET_TOKEN in the script with an actual token from the email/database"
curl -s -X POST "${BASE_URL}/reset-password" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "RESET_TOKEN",
    "password": "newpassword123"
  }'

# 5. Test Login with New Password
echo -e "\n${GREEN}5. Testing Login with New Password${NC}"
curl -s -X POST "${BASE_URL}/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "newpassword123"
  }'

# 6. Get Profile
echo -e "\n${GREEN}6. Testing Get Profile${NC}"
curl -s -X GET "${BASE_URL}/profile" \
  -H "Authorization: Bearer $TOKEN"

# 7. Update Profile
echo -e "\n${GREEN}7. Testing Update Profile${NC}"
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

# 8. Update Notification Preferences
echo -e "\n${GREEN}8. Testing Update Notification Preferences${NC}"
curl -s -X PUT "${BASE_URL}/profile/notifications" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email_notifications": true,
    "order_updates": true,
    "promotional_emails": false,
    "security_alerts": true
  }'

# 9. Add Address
echo -e "\n${GREEN}9. Testing Add Address${NC}"
curl -s -X POST "${BASE_URL}/addresses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "street": "123 Main St",
    "city": "New York",
    "state": "NY",
    "country": "USA",
    "postal_code": "10001",
    "is_default": true
  }'

# 10. List Addresses
echo -e "\n${GREEN}10. Testing List Addresses${NC}"
curl -s -X GET "${BASE_URL}/addresses" \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n${GREEN}✅ API tests completed${NC}" 