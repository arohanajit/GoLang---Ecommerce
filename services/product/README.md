# Product Service

## Database Setup

This service requires PostgreSQL 16. Make sure you have PostgreSQL installed and running.

### Initial Database Setup

1. Connect to PostgreSQL as superuser:
```bash
/Library/PostgreSQL/16/bin/psql -U postgres
```

2. Create the database and user:
```sql
CREATE USER product_user WITH PASSWORD '1235813';
CREATE DATABASE product_db;
GRANT ALL PRIVILEGES ON DATABASE product_db TO product_user;
```

### Environment Variables

Make sure your `.env` file contains the following configuration:
```
DB_HOST=localhost
DB_USER=product_user
DB_PASSWORD=1235813
DB_NAME=product_db
DB_PORT=5432
PORT=8000
CONSUL_HTTP_ADDR=http://localhost:8500
HOST_IP=127.0.0.1
```

### Troubleshooting

If you encounter database connection issues:

1. Verify PostgreSQL is running:
```bash
ps aux | grep postgres
```

2. Check if you can connect to the database:
```bash
/Library/PostgreSQL/16/bin/psql -U product_user -d product_db -h localhost
```

3. Verify the password is correct:
```bash
/Library/PostgreSQL/16/bin/psql -U postgres
ALTER USER product_user WITH PASSWORD '1235813';
```

4. Make sure the database exists and user has proper permissions:
```bash
/Library/PostgreSQL/16/bin/psql -U postgres -c "\l" | grep product_db
/Library/PostgreSQL/16/bin/psql -U postgres -c "\du" | grep product_user
``` 