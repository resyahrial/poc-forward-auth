# Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                            CLIENT                                    │
│                                                                       │
│   curl http://localhost/books                                        │
│   -H "Authorization: Bearer user123:admin:secret-token"              │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                │ (1) Request with Authorization header
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      TRAEFIK PROXY (Port 80)                         │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │         Router: book-service                                 │   │
│  │         Rule: PathPrefix(`/books`)                           │   │
│  │         Middleware: auth-middleware                          │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │   Middleware: auth-middleware (ForwardAuth)                  │   │
│  │   - address: http://auth-service:8081/auth                   │   │
│  │   - authResponseHeaders:                                     │   │
│  │     * X-User-ID                                              │   │
│  │     * X-User-Role                                            │   │
│  │     * X-Auth-Time                                            │   │
│  └─────────────────────────────────────────────────────────────┘   │
└───────────┬──────────────────────────────────┬──────────────────────┘
            │                                  │
            │ (2) Forward to auth              │ (5) Forward original request
            │     with X-Forwarded-* headers   │     with enriched headers
            │                                  │
            ▼                                  ▼
┌──────────────────────────────┐   ┌──────────────────────────────────┐
│   AUTH SERVICE (Port 8081)   │   │   BOOK SERVICE (Port 8082)        │
│                              │   │                                    │
│  Endpoints:                  │   │  Endpoints:                        │
│  - /auth (ForwardAuth)       │   │  - /books                          │
│  - /access-logs              │   │  - /health                         │
│  - /health                   │   │                                    │
│                              │   │  Functionality:                    │
│  Functionality:              │   │  1. Read enriched headers:         │
│  1. Validate token           │   │     - X-User-ID                    │
│  2. Extract user info        │   │     - X-User-Role                  │
│  3. Log access attempt       │   │     - X-Auth-Time                  │
│  4. Return 200 OK or 401     │   │  2. Filter books by role:          │
│  5. Set response headers:    │   │     - admin: all books             │
│     - X-User-ID              │   │     - user: public + user books    │
│     - X-User-Role            │   │     - guest: public books only     │
│     - X-Auth-Time            │   │  3. Return filtered data           │
└──────────────┬───────────────┘   └────────────────┬───────────────────┘
               │                                    │
               │ (3) 200 OK with headers            │ (6) Book data response
               │     X-User-ID: user123             │
               │     X-User-Role: admin             │
               │     X-Auth-Time: 2025-12-17...     │
               │                                    │
               └────────────────┬───────────────────┘
                                │
                                │ (4) Traefik copies headers
                                │     and forwards to book-service
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      TRAEFIK PROXY                                   │
│                                                                       │
│  Copies headers from auth response and adds to forwarded request     │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                │ (7) Final response to client
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                            CLIENT                                    │
│                                                                       │
│   Response:                                                          │
│   {                                                                  │
│     "user_id": "user123",                                            │
│     "role": "admin",                                                 │
│     "total_books": 6,                                                │
│     "books": [...]                                                   │
│   }                                                                  │
└─────────────────────────────────────────────────────────────────────┘
```

## Request Flow Details

### Successful Authentication Flow

1. **Client Request**: Client sends GET request to `http://localhost/books` with `Authorization` header
2. **Traefik Intercept**: Traefik matches the request to `book-service` router and applies `auth-middleware`
3. **Forward to Auth**: Traefik forwards the request to auth service with:
   - Original `Authorization` header
   - X-Forwarded-Method: GET
   - X-Forwarded-Uri: /books
   - X-Forwarded-For: client IP
   - X-Forwarded-Host: localhost
4. **Token Validation**: Auth service validates the Bearer token
5. **Access Logging**: Auth service records the access attempt
6. **Response with Headers**: Auth service returns 200 OK with custom headers:
   - X-User-ID: extracted from token
   - X-User-Role: extracted from token
   - X-Auth-Time: current timestamp
7. **Header Copy**: Traefik copies the specified headers from auth response
8. **Forward to Backend**: Traefik forwards original request to book service with enriched headers
9. **Filter Books**: Book service reads role from X-User-Role header and filters books accordingly
10. **Response**: Book service returns filtered book data
11. **Client Response**: Traefik forwards the response back to client

### Failed Authentication Flow

1. **Client Request**: Client sends request without valid token
2. **Traefik Intercept**: Traefik applies auth-middleware
3. **Forward to Auth**: Traefik forwards to auth service
4. **Validation Fails**: Auth service validates and finds invalid/missing token
5. **Return 401**: Auth service returns 401 Unauthorized
6. **Client Response**: Traefik forwards 401 response directly to client (book service is never called)

## Key Components

### Traefik Configuration

**Static Config** (`traefik.yml`):
- Defines entry points
- Enables dashboard
- Configures file provider

**Dynamic Config** (`dynamic-config.yml`):
- Defines routers (book-service, auth-service)
- Defines middleware (auth-middleware)
- Defines services (load balancers)

### Auth Service Features

- **Token Format**: `Bearer <userID>:<role>:<token>`
- **Valid Token**: `secret-token`
- **Response Headers**: X-User-ID, X-User-Role, X-Auth-Time
- **Access Logging**: In-memory storage of all requests

### Book Service Features

- **6 Books Total**:
  - 2 public books
  - 2 user books
  - 2 admin books
- **Role-Based Filtering**:
  - admin sees all 6
  - user sees 4 (public + user)
  - guest sees 2 (public only)

## Network Architecture

```
┌────────────────────────────────────────┐
│         Docker Network: poc-network    │
│                                        │
│  ┌──────────┐  ┌────────────┐  ┌────────────┐
│  │ Traefik  │  │    Auth    │  │    Book    │
│  │ :80,:8080│  │  Service   │  │  Service   │
│  │          │──│   :8081    │  │   :8082    │
│  └──────────┘  └────────────┘  └────────────┘
│       │                                │
└───────┼────────────────────────────────┼──────┘
        │                                │
        │                                │
   Host :80                         Host :8082
   Host :8080                       Host :8081
   (Traefik)                        (Direct Access)
```

## Security Flow

```
Request → Traefik → Auth Validation → Approved?
                         │               │
                         │               ├─ Yes → Enrich Headers → Book Service
                         │               │
                         │               └─ No → Return 401
                         │
                         └─ Log Access Attempt
```
