# Traefik Forward Auth POC

A proof of concept demonstrating Traefik's Forward Auth middleware with two microservices: an authentication service and a protected book service.

## Table of Contents
- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Token Format & Access Levels](#token-format--access-levels)
- [Usage Examples](#usage-examples)
- [API Reference](#api-reference)
- [Make Commands](#make-commands)
- [Troubleshooting](#troubleshooting)
- [Security Notes](#security-notes)

## Quick Start

### 1. Start Services

```bash
# Start services
docker-compose up -d
# or
make up
```

Wait ~5 seconds for services to initialize.

### 2. Test Authentication

```bash
# Admin (sees all 6 books)
curl http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token"

# User (sees 4 books: public + user)
curl http://localhost/books \
  -H "Authorization: Bearer alice:user:secret-token"

# Guest (sees 2 books: public only)
curl http://localhost/books \
  -H "Authorization: Bearer bob:guest:secret-token"

# No auth (fails with 401)
curl http://localhost/books
```

### 3. View Access Logs

```bash
curl http://localhost/access-logs | jq
```

### 4. Traefik Dashboard

Open: http://localhost:8080

### 5. Stop Services

```bash
docker-compose down
# or
make down
```

## Architecture

### System Overview

```
Client
  ↓
[Traefik :80]
  ↓
  ├─→ [Auth Service :8081] ──→ Validate Token
  │                            ↓
  │                    Enrich Headers (X-User-ID, X-User-Role, X-Auth-Time)
  │                            ↓
  └─→ [Book Service :8082] ──→ Filter Books by Role
                               ↓
                        Return Filtered Data
```

**Flow:** Client → Traefik → Auth Service (validates) → Book Service (filtered data)

**Services:**
- **Traefik** (Port 80, 8080) - Reverse proxy with forward auth
- **Auth Service** (Port 8081) - Token validation, header enrichment, access logging
- **Book Service** (Port 8082) - Role-based book filtering

### Request Flow

#### Successful Authentication

1. **Client Request**
   ```
   GET /books
   Authorization: Bearer admin:admin:secret-token
   ```

2. **Traefik Forward Auth**
   - Intercepts request
   - Forwards to `auth-service:8081/auth`
   - Adds X-Forwarded-* headers

3. **Auth Service**
   - Validates token (`secret-token`)
   - Logs access attempt
   - Returns 200 OK with headers:
     - `X-User-ID: admin`
     - `X-User-Role: admin`
     - `X-Auth-Time: 2025-12-17T...`

4. **Traefik Header Copy**
   - Copies enriched headers from auth response
   - Forwards original request to book service

5. **Book Service**
   - Reads `X-User-Role` header
   - Filters books by role
   - Returns filtered data

6. **Response**
   ```json
   {
     "user_id": "admin",
     "role": "admin",
     "total_books": 6,
     "books": [...]
   }
   ```

#### Failed Authentication

1. Client sends invalid/missing token
2. Traefik forwards to auth service
3. Auth service returns 401
4. Traefik returns 401 to client (book service never called)

### Components

#### Traefik

**Static Config** (`traefik.yml`):
- Entry points (port 80)
- Dashboard (port 8080)
- File provider

**Dynamic Config** (`dynamic-config.yml`):
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

**Key Configuration Options:**
- **address**: URL of the authentication service
- **authResponseHeaders**: Headers to copy from auth response to forwarded request
- **preserveRequestMethod**: Preserve original HTTP method when forwarding to auth service

#### Auth Service

**Functions:**
- Token validation (Bearer format)
- Access logging (in-memory)
- Header enrichment

**Endpoints:**
- `POST /auth` - ForwardAuth endpoint
- `GET /access-logs` - View logs
- `GET /health` - Health check

#### Book Service

**Functions:**
- Role-based filtering
- 6 books: 2 public, 2 user, 2 admin

**Access Matrix:**
| Role  | Public (2) | User (2) | Admin (2) | Total |
|-------|-----------|----------|-----------|-------|
| admin | ✅        | ✅       | ✅        | 6     |
| user  | ✅        | ✅       | ❌        | 4     |
| other | ✅        | ❌       | ❌        | 2     |

**Endpoints:**
- `GET /books` - Get filtered books
- `GET /health` - Health check

### Network Architecture

```
Docker Network: poc-network
  │
  ├─ Traefik
  │   Ports: 80 (proxy), 8080 (dashboard)
  │
  ├─ Auth Service
  │   Port: 8081
  │   Internal: auth-service:8081
  │
  └─ Book Service
      Port: 8082
      Internal: book-service:8082
```

### Technologies

- **Traefik v3.0** - Reverse proxy
- **Go 1.23** - Services language
- **Docker** - Containerization
- **Alpine Linux** - Base image

## Project Structure

```
poc-forward-auth/
├── auth-service/           # Authentication service (Go)
│   ├── main.go            # Token validation, logging
│   ├── main_test.go       # Unit tests (72.3% coverage)
│   ├── Dockerfile
│   └── go.mod
├── book-service/          # Book service (Go)
│   ├── main.go           # Role-based filtering
│   ├── main_test.go      # Unit tests (59.4% coverage)
│   ├── Dockerfile
│   └── go.mod
├── traefik/
│   ├── traefik.yml       # Static config
│   └── dynamic-config.yml # Routes & middleware
├── docker-compose.yml     # Service orchestration
├── Makefile              # Convenient commands
├── test.sh               # Integration tests
└── run-tests.sh          # Unit tests
```

## Token Format & Access Levels

**Format:** `Authorization: Bearer <userID>:<role>:secret-token`

**Valid token:** `secret-token`

**Examples:**
- Admin: `Bearer admin:admin:secret-token` → 6 books (all)
- User: `Bearer alice:user:secret-token` → 4 books (public + user)
- Guest: `Bearer bob:guest:secret-token` → 2 books (public only)

## Testing

### Unit Tests

```bash
# Run all unit tests
make test-unit
# or
./run-tests.sh

# Auth service only
make test-auth

# Book service only
make test-book
```

### Integration Tests

```bash
# Run integration tests with Docker
make test
# or
./test.sh
```

## Usage Examples

### Authentication Tests

#### ✅ Valid Admin Token
```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token"
```
**Expected**: 200 OK, returns all 6 books

#### ✅ Valid User Token
```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer alice:user:secret-token"
```
**Expected**: 200 OK, returns 4 books (public + user)

#### ✅ Valid Guest Token
```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer bob:guest:secret-token"
```
**Expected**: 200 OK, returns 2 books (public only)

#### ❌ Invalid Token
```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer user:admin:wrong-token"
```
**Expected**: 401 Unauthorized

#### ❌ Missing Token
```bash
curl http://localhost/books
```
**Expected**: 401 Unauthorized

#### ❌ Malformed Token
```bash
curl -v http://localhost/books \
  -H "Authorization: InvalidFormat"
```
**Expected**: 401 Unauthorized

### Different User Scenarios

#### Scenario A: Admin accessing books
```bash
curl -s http://localhost/books \
  -H "Authorization: Bearer superadmin:admin:secret-token" | jq
```
**Result**:
```json
{
  "access_level": "admin",
  "auth_time": "2025-12-17T...",
  "books": [ /* all 6 books */ ],
  "role": "admin",
  "total_books": 6,
  "user_id": "superadmin"
}
```

#### Scenario B: Regular user accessing books
```bash
curl -s http://localhost/books \
  -H "Authorization: Bearer john:user:secret-token" | jq
```
**Result**:
```json
{
  "access_level": "user",
  "auth_time": "2025-12-17T...",
  "books": [ /* 4 books: public + user */ ],
  "role": "user",
  "total_books": 4,
  "user_id": "john"
}
```

#### Scenario C: Guest accessing books
```bash
curl -s http://localhost/books \
  -H "Authorization: Bearer visitor:guest:secret-token" | jq
```
**Result**:
```json
{
  "access_level": "guest",
  "auth_time": "2025-12-17T...",
  "books": [ /* 2 books: public only */ ],
  "role": "guest",
  "total_books": 2,
  "user_id": "visitor"
}
```

### Access Logging

#### View all access logs
```bash
curl -s http://localhost/access-logs | jq
```
**Result**:
```json
{
  "logs": [
    {
      "forwarded_for": "172.18.0.1",
      "method": "GET",
      "role": "admin",
      "timestamp": "2025-12-17T10:30:00Z",
      "uri": "/books",
      "user_id": "admin"
    }
  ],
  "total_requests": 1
}
```

#### Count total requests
```bash
curl -s http://localhost/access-logs | jq '.total_requests'
```

#### Filter logs by user
```bash
curl -s http://localhost/access-logs | jq '.logs[] | select(.user_id == "admin")'
```

### Health Checks

```bash
# Auth service health
curl http://localhost:8081/health

# Book service health
curl http://localhost:8082/health
```

### Response Analysis

#### Get only book count
```bash
curl -s http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token" | jq '.total_books'
```

#### Get only book titles
```bash
curl -s http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token" | jq '.books[].title'
```

#### Get books by access level
```bash
# Public books
curl -s http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token" | jq '.books[] | select(.access == "public")'

# User books
curl -s http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token" | jq '.books[] | select(.access == "user")'

# Admin books
curl -s http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token" | jq '.books[] | select(.access == "admin")'
```

### Performance Testing

#### Sequential requests
```bash
for i in {1..10}; do
  curl -s http://localhost/books \
    -H "Authorization: Bearer user$i:user:secret-token" | jq '.user_id'
done
```

#### Check access log count after multiple requests
```bash
# Make 5 requests
for i in {1..5}; do
  curl -s http://localhost/books \
    -H "Authorization: Bearer user$i:admin:secret-token" > /dev/null
done

# Check total count
curl -s http://localhost/access-logs | jq '.total_requests'
```

### Comparison Tests

#### Compare book count across roles
```bash
echo "Admin books:"
curl -s http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token" | jq '.total_books'

echo "User books:"
curl -s http://localhost/books \
  -H "Authorization: Bearer user:user:secret-token" | jq '.total_books'

echo "Guest books:"
curl -s http://localhost/books \
  -H "Authorization: Bearer guest:guest:secret-token" | jq '.total_books'
```

### Quick Test Sequence

Run these commands in order to test the complete flow:

```bash
# 1. Start services
make up

# 2. Wait for services to start
sleep 5

# 3. Test without auth (should fail)
curl -v http://localhost/books

# 4. Test with admin (should see 6 books)
curl -s http://localhost/books -H "Authorization: Bearer admin:admin:secret-token" | jq '.total_books'

# 5. Test with user (should see 4 books)
curl -s http://localhost/books -H "Authorization: Bearer user:user:secret-token" | jq '.total_books'

# 6. Test with guest (should see 2 books)
curl -s http://localhost/books -H "Authorization: Bearer guest:guest:secret-token" | jq '.total_books'

# 7. Check access logs (should show 3 successful requests)
curl -s http://localhost/access-logs | jq '.total_requests'

# 8. Open Traefik dashboard
open http://localhost:8080
```

## API Reference

### Ports

- **80** - Traefik (main proxy)
- **8080** - Traefik Dashboard
- **8081** - Auth Service (direct access)
- **8082** - Book Service (direct access)

### Auth Service Endpoints

#### POST /auth (ForwardAuth Endpoint)
- **Purpose**: Validate authentication token
- **Headers**: Authorization: Bearer <userID>:<role>:<token>
- **Success Response**: 200 OK with enriched headers
  - X-User-ID: <userID>
  - X-User-Role: <role>
  - X-Auth-Time: <timestamp>
- **Failure Response**: 401 Unauthorized

#### GET /access-logs
- **Purpose**: View access logs
- **Response**: JSON with logs array and total_requests count

#### GET /health
- **Purpose**: Health check
- **Response**: "Auth service is healthy"

### Book Service Endpoints

#### GET /books
- **Purpose**: Get filtered books based on role
- **Headers** (from Traefik):
  - X-User-ID: <userID>
  - X-User-Role: <role>
  - X-Auth-Time: <timestamp>
- **Response**: JSON with user info and filtered books

#### GET /health
- **Purpose**: Health check
- **Response**: "Book service is healthy"

### Traefik Dashboard

#### Access
- **URL**: http://localhost:8080
- **Features**: View routers, middlewares, services

#### API Endpoints (if enabled)
```bash
# Get routers
curl http://localhost:8080/api/http/routers

# Get middlewares
curl http://localhost:8080/api/http/middlewares

# Get services
curl http://localhost:8080/api/http/services
```

## Make Commands

```bash
make up           # Start services
make down         # Stop services
make logs         # View all logs
make test         # Run integration tests
make test-unit    # Run unit tests
make test-all     # Run all tests
make test-auth    # Test auth service only
make test-book    # Test book service only
make admin        # Test admin access
make user         # Test user access
make guest        # Test guest access
make logs-auth    # View auth service logs
make logs-book    # View book service logs
make logs-traefik # View Traefik logs
```

## Troubleshooting

### Services won't start

```bash
# Check ports
lsof -i :80 :8080 :8081 :8082

# Rebuild
docker-compose down && docker-compose up --build
```

### Authentication fails

- Verify token format: `Bearer <userID>:<role>:secret-token`
- Token must be exactly: `secret-token`
- Check auth service logs: `make logs-auth`

### Books not filtered

- Check enriched headers in book service logs
- Verify Traefik config: `authResponseHeaders` in `dynamic-config.yml`

### Docker logs

```bash
# View all logs
docker-compose logs -f

# View auth service logs only
docker-compose logs -f auth-service

# View book service logs only
docker-compose logs -f book-service

# View Traefik logs only
docker-compose logs -f traefik
```

## Security Notes

⚠️ **This is a POC for demonstration only. For production:**

- Use proper JWT/OAuth2 tokens instead of simple strings
- Enable HTTPS/TLS for all communications
- Don't expose services directly (only through Traefik)
- Use secret management for tokens
- Implement rate limiting
- Add proper logging and monitoring
- Protect accessLogs with mutex for concurrent access
- Add authentication to Traefik dashboard
- Implement proper error handling
- Use database for access logs instead of in-memory storage
- Add request validation and sanitization
- Implement CORS policies
- Add audit logging

## Resources

- [Traefik Forward Auth Documentation](https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/forwardauth/)
- [Docker Documentation](https://docs.docker.com/)
- [Go Documentation](https://golang.org/doc/)

---

**Note**: This is a proof of concept for educational purposes. Do not use in production without implementing proper security measures.
