#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to kill service by port
kill_service() {
    local port=$1
    local pid=$(lsof -ti :$port)
    if [ ! -z "$pid" ]; then
        echo -e "${RED}Killing process on port $port (PID: $pid)${NC}"
        kill -9 $pid 2>/dev/null
        echo -e "${GREEN}Service on port $port stopped${NC}"
    else
        echo -e "${GREEN}No service running on port $port${NC}"
    fi
}

echo "Stopping all services..."

# Stop each service
kill_service 8081  # Gateway
kill_service 8002  # User Service
kill_service 8000  # Product Service
kill_service 8001  # Order Service
kill_service 8003  # Inventory Service

echo -e "${GREEN}All services have been stopped!${NC}"