# E-Commerce Platform

A modern, microservices-based e-commerce platform built with Go, featuring service discovery, containerization, and a robust API Gateway.

## Architecture Overview

The platform consists of the following microservices:

- **API Gateway** (Port 8081)
  - Entry point for all client requests
  - Handles service discovery and request routing
  - Implements health checks and load balancing

- **Product Service** (Port 8000)
  - Manages product catalog
  - Handles CRUD operations for products
  - Supports product search and filtering

- **User Service** (Port 8002)
  - Manages user accounts and authentication
  - Handles user profile management
  - Supports address management
  - Implements password reset functionality

- **Order Service** (Port 8001)
  - Processes and manages orders
  - Handles order status updates
  - Maintains order history

- **Inventory Service** (Port 8003)
  - Manages product stock levels
  - Handles stock updates
  - Tracks inventory transactions
  - Implements stock alerts

- **Payment Service** (Port 8004)
  - Processes payments
  - Handles payment status updates
  - Supports multiple payment methods

## Technology Stack

- **Backend**: Go (1.22)
- **Web Framework**: Gin
- **Database**: PostgreSQL 15
- **Service Discovery**: Consul
- **Containerization**: Docker
- **API Documentation**: OpenAPI/Swagger
- **Authentication**: JWT

## Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose
- PostgreSQL 15
- Git

## Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/e-commerce-platform.git
   cd e-commerce-platform
   ```

2. Set up environment variables for each service. Create `.env` files in each service directory:

   **Gateway Service (.env)**:
   ```
   PORT=8081
   CONSUL_HTTP_ADDR=http://consul:8500
   ```

   **Product Service (.env)**:
   ```
   DB_HOST=postgres
   DB_USER=product_user
   DB_PASSWORD=1235813
   DB_NAME=product_db
   DB_PORT=5432
   PORT=8000
   CONSUL_HTTP_ADDR=http://consul:8500
   ```

   **User Service (.env)**:
   ```
   DB_HOST=postgres
   DB_USER=user_user
   DB_PASSWORD=1235813
   DB_NAME=user_db
   DB_PORT=5432
   PORT=8002
   CONSUL_HTTP_ADDR=http://consul:8500
   JWT_SECRET=your-secret-key
   ```

   **Order Service (.env)**:
   ```
   DB_HOST=postgres
   DB_USER=order_user
   DB_PASSWORD=1235813
   DB_NAME=order_db
   DB_PORT=5432
   PORT=8001
   CONSUL_HTTP_ADDR=http://consul:8500
   ```

   **Inventory Service (.env)**:
   ```
   DB_HOST=postgres
   DB_USER=inventory_user
   DB_PASSWORD=1235813
   DB_NAME=inventory_db
   DB_PORT=5432
   PORT=8003
   CONSUL_HTTP_ADDR=http://consul:8500
   ```

   **Payment Service (.env)**:
   ```
   DB_HOST=postgres
   DB_USER=payment_user
   DB_PASSWORD=1235813
   DB_NAME=payment_db
   DB_PORT=5432
   PORT=8004
   CONSUL_HTTP_ADDR=http://consul:8500
   ```

3. Start the services using Docker Compose:
   ```bash
   docker compose up
   ```

## API Documentation

### Product Service

- `GET /api/v1/products` - List all products
- `POST /api/v1/products` - Create a new product
- `GET /api/v1/products/:id` - Get product details
- `PUT /api/v1/products/:id` - Update product
- `DELETE /api/v1/products/:id` - Delete product

### User Service

- `POST /register` - Register new user
- `POST /login` - User login
- `POST /forgot-password` - Request password reset
- `POST /reset-password` - Reset password
- `GET /profile` - Get user profile
- `PUT /profile` - Update user profile
- `PUT /profile/change-password` - Change password
- `DELETE /profile` - Delete account
- `POST /addresses` - Add address
- `GET /addresses` - List addresses
- `PUT /addresses/:id` - Update address
- `DELETE /addresses/:id` - Delete address

### Order Service

- `GET /api/v1/orders` - List orders
- `POST /api/v1/orders` - Create order
- `GET /api/v1/orders/:id` - Get order details
- `PUT /api/v1/orders/:id` - Update order
- `DELETE /api/v1/orders/:id` - Delete order

### Inventory Service

- `POST /api/v1/inventory/items` - Create inventory item
- `GET /api/v1/inventory/items` - List inventory
- `GET /api/v1/inventory/items/:id` - Get item details
- `PUT /api/v1/inventory/items/:id/stock` - Update stock
- `GET /api/v1/inventory/items/:id/transactions` - Get transaction history

### Payment Service

- `POST /api/v1/payments` - Process payment
- `GET /api/v1/payments/:id` - Get payment details

## Service Discovery and Health Checks

All services register with Consul for service discovery. Health checks are configured with:
- 10-second intervals
- 1-second timeouts
- 30-second deregistration for critical services

## Database Schema

Each service maintains its own database:

### Product Service
```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR NOT NULL,
    description TEXT,
    price DECIMAL NOT NULL,
    stock INTEGER NOT NULL,
    images TEXT[],
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### User Service
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR UNIQUE NOT NULL,
    password_hash VARCHAR NOT NULL,
    first_name VARCHAR,
    last_name VARCHAR,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    street VARCHAR NOT NULL,
    city VARCHAR NOT NULL,
    state VARCHAR NOT NULL,
    postal_code VARCHAR NOT NULL,
    country VARCHAR NOT NULL,
    is_default BOOLEAN DEFAULT false
);
```

## Development

### Building Services Locally

Each service can be built and run independently:

```bash
cd services/product
go build
./product-service
```

### Running Tests

```bash
go test ./...
```

### Code Style

The project follows standard Go code style guidelines. Format your code using:

```bash
go fmt ./...
```

## Deployment

The platform is containerized using Docker and can be deployed using Docker Compose or Kubernetes.

### Docker Compose Deployment

```bash
docker compose up -d
```

### Scaling Services

```bash
docker compose up -d --scale product-service=3
```

## Monitoring and Logging

- All services implement structured logging
- Health endpoints provide service status
- Consul UI available at http://localhost:8500

## Troubleshooting

Common issues and solutions:

1. **Database Connection Issues**
   - Verify PostgreSQL is running
   - Check database credentials
   - Ensure database exists and user has proper permissions

2. **Service Discovery Issues**
   - Verify Consul is running
   - Check service registration logs
   - Ensure correct Consul address in environment variables

3. **API Gateway Issues**
   - Check service health in Consul
   - Verify service addresses and ports
   - Check gateway logs for routing errors

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
