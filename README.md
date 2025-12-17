# Traefik Forward Auth POC

This is a proof of concept (POC) demonstrating Traefik's Forward Auth middleware feature. The implementation includes two services: an authentication service and a book service, where requests to the book service are intercepted and authenticated by the auth service using Traefik's forward auth middleware.

## Architecture

```
Client Request
     |
     v
[Traefik Proxy]
     |
     |---> [Auth Service] (validates token, enriches headers, logs access)
     |           |
     |           v
     |      (returns 200 OK with headers)
     |
     v
[Book Service] (receives enriched headers, returns filtered data)
```

## Components

### 1. Auth Service (Port 8081)
- **Purpose**: Validates access tokens and authenticates requests
- **Features**:
  - Token validation (Bearer token format)
  - Request header enrichment (adds user ID, role, and auth time)
  - Access logging (records all authentication attempts)
  
- **Endpoints**:
  - `POST /auth` - Forward auth endpoint (called by Traefik)
  - `GET /access-logs` - View all access logs
  - `GET /health` - Health check

### 2. Book Service (Port 8082)
- **Purpose**: Protected resource that returns books based on user access level
- **Features**:
  - Role-based access control (public, user, admin)
  - Reads enriched headers from auth service
  - Returns filtered book data based on user role

- **Endpoints**:
  - `GET /books` - Get books (requires authentication)
  - `GET /health` - Health check

### 3. Traefik (Port 80, Dashboard 8080)
- **Purpose**: Reverse proxy with forward auth middleware
- **Features**:
  - Routes requests to appropriate services
  - Intercepts book service requests for authentication
  - Forwards enriched headers to backend services
  - Dashboard for monitoring

## Token Format

The auth service expects tokens in the following format:

```
Authorization: Bearer <userID>:<role>:<token>
```

**Examples**:
- `Authorization: Bearer user123:admin:secret-token`
- `Authorization: Bearer alice:user:secret-token`
- `Authorization: Bearer bob:guest:secret-token`

**Valid token**: `secret-token`

## Access Levels

| Role | Public Books | User Books | Admin Books |
|------|--------------|------------|-------------|
| admin | ✅ | ✅ | ✅ |
| user | ✅ | ✅ | ❌ |
| guest | ✅ | ❌ | ❌ |
| unknown | ✅ | ❌ | ❌ |

## Testing

### Unit Tests

Run unit tests for both services:

```bash
make test-unit
# or
./run-tests.sh
```

Run tests for individual services:

```bash
# Auth service only
make test-auth

# Book service only
make test-book
```

**Test Coverage:**
- Auth Service: 72.3%
- Book Service: 59.4%

See [TESTING.md](TESTING.md) for detailed test documentation.

### Integration Tests

Run integration tests with Docker:

```bash
make test
# or
./test.sh
```

### All Tests

Run both unit and integration tests:

```bash
make test-all
```

## Setup and Run

### Prerequisites
- Docker
- Docker Compose
- Go 1.23+ (for running tests locally)

### Start the Services

```bash
# Build and start all services
docker-compose up --build

# Or run in detached mode
docker-compose up --build -d
```

### Stop the Services

```bash
docker-compose down
```

## Usage Examples

### 1. Request without Authentication (should fail)

```bash
curl -v http://localhost/books
```

**Expected Response**: `401 Unauthorized`

### 2. Request with Admin Token

```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token"
```

**Expected Response**: All books (public, user, and admin books)

```json
{
  "access_level": "admin",
  "auth_time": "2025-12-17T10:30:00Z",
  "books": [
    {"id": "1", "title": "Public Book 1", "author": "Author A", "access": "public"},
    {"id": "2", "title": "Public Book 2", "author": "Author B", "access": "public"},
    {"id": "3", "title": "User Book 1", "author": "Author C", "access": "user"},
    {"id": "4", "title": "User Book 2", "author": "Author D", "access": "user"},
    {"id": "5", "title": "Admin Book 1", "author": "Author E", "access": "admin"},
    {"id": "6", "title": "Admin Book 2", "author": "Author F", "access": "admin"}
  ],
  "role": "admin",
  "total_books": 6,
  "user_id": "admin"
}
```

### 3. Request with User Token

```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer alice:user:secret-token"
```

**Expected Response**: Public and user books only

```json
{
  "access_level": "user",
  "auth_time": "2025-12-17T10:31:00Z",
  "books": [
    {"id": "1", "title": "Public Book 1", "author": "Author A", "access": "public"},
    {"id": "2", "title": "Public Book 2", "author": "Author B", "access": "public"},
    {"id": "3", "title": "User Book 1", "author": "Author C", "access": "user"},
    {"id": "4", "title": "User Book 2", "author": "Author D", "access": "user"}
  ],
  "role": "user",
  "total_books": 4,
  "user_id": "alice"
}
```

### 4. Request with Guest Token

```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer bob:guest:secret-token"
```

**Expected Response**: Public books only

```json
{
  "access_level": "guest",
  "auth_time": "2025-12-17T10:32:00Z",
  "books": [
    {"id": "1", "title": "Public Book 1", "author": "Author A", "access": "public"},
    {"id": "2", "title": "Public Book 2", "author": "Author B", "access": "public"}
  ],
  "role": "guest",
  "total_books": 2,
  "user_id": "bob"
}
```

### 5. View Access Logs

```bash
curl http://localhost/access-logs
```

**Expected Response**: List of all authentication attempts

```json
{
  "logs": [
    {
      "timestamp": "2025-12-17T10:30:00Z",
      "method": "GET",
      "uri": "/books",
      "forwarded_for": "172.18.0.1",
      "user_id": "admin",
      "role": "admin"
    },
    {
      "timestamp": "2025-12-17T10:31:00Z",
      "method": "GET",
      "uri": "/books",
      "forwarded_for": "172.18.0.1",
      "user_id": "alice",
      "role": "user"
    }
  ],
  "total_requests": 2
}
```

### 6. Direct Access to Auth Service (bypassing Traefik)

```bash
# Health check
curl http://localhost:8081/health

# Access logs
curl http://localhost:8081/access-logs
```

### 7. Direct Access to Book Service (bypassing Traefik - no auth)

```bash
curl http://localhost:8082/books
```

**Note**: Direct access bypasses authentication. In production, services should only be accessible through Traefik.

## Traefik Dashboard

Access the Traefik dashboard at:
```
http://localhost:8080
```

Here you can monitor:
- Active routers
- Configured middlewares
- Services status
- Request metrics

## How Forward Auth Works

1. **Client Request**: Client sends a request to `http://localhost/books` with Authorization header
2. **Traefik Intercept**: Traefik intercepts the request and applies the `auth-middleware`
3. **Forward to Auth**: Traefik forwards the request to `http://auth-service:8081/auth` with X-Forwarded-* headers
4. **Authentication**: Auth service validates the token:
   - If invalid: returns 401, Traefik returns 401 to client
   - If valid: returns 200 with enriched headers (X-User-ID, X-User-Role, X-Auth-Time)
5. **Access Logging**: Auth service logs the access attempt
6. **Header Enrichment**: Traefik copies the headers from auth response
7. **Forward to Backend**: Traefik forwards the original request to book service with enriched headers
8. **Response**: Book service filters books based on role and returns the response

## Configuration Details

### Traefik Dynamic Configuration

The forward auth middleware is configured in `traefik/dynamic-config.yml`:

```yaml
middlewares:
  auth-middleware:
    forwardAuth:
      address: "http://auth-service:8081/auth"
      authResponseHeaders:
        - "X-User-ID"
        - "X-User-Role"
        - "X-Auth-Time"
      preserveRequestMethod: true
```

### Key Configuration Options

- **address**: URL of the authentication service
- **authResponseHeaders**: Headers to copy from auth response to forwarded request
- **preserveRequestMethod**: Preserve original HTTP method when forwarding to auth service

## Testing Different Scenarios

### Test Invalid Token

```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer user:admin:wrong-token"
```

**Expected**: `401 Unauthorized`

### Test Missing Token

```bash
curl -v http://localhost/books
```

**Expected**: `401 Unauthorized`

### Test Different Roles

```bash
# Admin
curl http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token" | jq '.total_books'
# Output: 6

# User
curl http://localhost/books \
  -H "Authorization: Bearer user:user:secret-token" | jq '.total_books'
# Output: 4

# Guest
curl http://localhost/books \
  -H "Authorization: Bearer guest:guest:secret-token" | jq '.total_books'
# Output: 2
```

## Logs

### View Auth Service Logs

```bash
docker-compose logs -f auth-service
```

### View Book Service Logs

```bash
docker-compose logs -f book-service
```

### View Traefik Logs

```bash
docker-compose logs -f traefik
```

## Security Considerations

This is a POC for demonstration purposes. In production:

1. **Use proper token validation**: Implement JWT or OAuth2 tokens
2. **Use HTTPS**: Enable TLS/SSL for all communications
3. **Secure the services**: Don't expose services directly, only through Traefik
4. **Store secrets securely**: Use environment variables or secret management systems
5. **Implement rate limiting**: Protect against brute force attacks
6. **Add proper logging**: Use structured logging and centralized log management
7. **Network segmentation**: Use Docker networks properly to isolate services

## Troubleshooting

### Services not starting

```bash
# Check if ports are available
lsof -i :80
lsof -i :8080
lsof -i :8081
lsof -i :8082

# Rebuild images
docker-compose build --no-cache
docker-compose up
```

### Authentication failing

- Check auth service logs: `docker-compose logs auth-service`
- Verify token format: `Bearer <userID>:<role>:secret-token`
- Ensure token is `secret-token`

### Books not filtered correctly

- Check if headers are being passed: Look at book service logs
- Verify Traefik configuration: Check `authResponseHeaders` in dynamic-config.yml

## License

This is a proof of concept for educational purposes.
