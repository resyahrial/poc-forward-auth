# Project Structure

```
poc-forward-auth/
├── README.md                    # Complete documentation
├── QUICKSTART.md               # Quick start guide
├── ARCHITECTURE.md             # Architecture diagrams and flow
├── EXAMPLES.md                 # API call examples
├── Makefile                    # Convenient make commands
├── docker-compose.yml          # Docker orchestration
├── test.sh                     # Automated test script
├── .gitignore                  # Git ignore rules
│
├── auth-service/               # Authentication Service
│   ├── main.go                 # Auth service implementation
│   ├── go.mod                  # Go module definition
│   └── Dockerfile              # Auth service container
│
├── book-service/               # Book Service (Protected Resource)
│   ├── main.go                 # Book service implementation
│   ├── go.mod                  # Go module definition
│   └── Dockerfile              # Book service container
│
└── traefik/                    # Traefik Configuration
    ├── traefik.yml             # Static configuration
    └── dynamic-config.yml      # Dynamic configuration (routes, middleware)
```

## File Descriptions

### Documentation Files

- **README.md**: Comprehensive documentation including architecture, setup, usage, and troubleshooting
- **QUICKSTART.md**: Quick start guide for getting up and running fast
- **ARCHITECTURE.md**: Detailed architecture diagrams and request flow explanations
- **EXAMPLES.md**: Extensive collection of curl command examples for testing

### Configuration Files

- **docker-compose.yml**: Orchestrates all three services (Traefik, Auth, Book)
- **Makefile**: Provides convenient commands (make up, make test, make logs, etc.)
- **test.sh**: Automated testing script that runs all test scenarios
- **.gitignore**: Excludes build artifacts, binaries, and IDE files

### Auth Service

- **main.go**: 
  - Token validation (Bearer token format)
  - Access logging functionality
  - Header enrichment (X-User-ID, X-User-Role, X-Auth-Time)
  - Endpoints: /auth, /access-logs, /health

- **go.mod**: Go module definition
- **Dockerfile**: Multi-stage build for minimal container size

### Book Service

- **main.go**:
  - Role-based book filtering
  - Reads enriched headers from auth service
  - Returns filtered book data
  - Endpoints: /books, /health

- **go.mod**: Go module definition
- **Dockerfile**: Multi-stage build for minimal container size

### Traefik Configuration

- **traefik.yml** (Static):
  - Entry points configuration (port 80)
  - Dashboard configuration (port 8080)
  - File provider setup
  - Logging configuration

- **dynamic-config.yml** (Dynamic):
  - HTTP routers (book-service, auth-service)
  - ForwardAuth middleware configuration
  - Service definitions
  - Header copying configuration

## Quick Reference

### Common Commands

```bash
# Start services
make up
# or
docker-compose up -d

# Run tests
make test
# or
./test.sh

# View logs
make logs

# Test with different roles
make admin    # See all 6 books
make user     # See 4 books (public + user)
make guest    # See 2 books (public only)

# View access logs
make access-logs

# Stop services
make down
```

### Service Endpoints

| Service | Port | Endpoint | Description |
|---------|------|----------|-------------|
| Traefik | 80 | /books | Protected book endpoint (via Traefik) |
| Traefik | 80 | /auth | Auth endpoint (via Traefik) |
| Traefik | 80 | /access-logs | Access logs (via Traefik) |
| Traefik | 8080 | / | Traefik dashboard |
| Auth | 8081 | /auth | ForwardAuth endpoint |
| Auth | 8081 | /access-logs | View access logs |
| Auth | 8081 | /health | Health check |
| Book | 8082 | /books | Get books |
| Book | 8082 | /health | Health check |

### Token Format

```
Authorization: Bearer <userID>:<role>:<token>
```

- **userID**: Any string (e.g., "user123", "alice", "admin")
- **role**: Any string (e.g., "admin", "user", "guest")
- **token**: Must be "secret-token" (this is the only valid token)

### Access Matrix

| Role | Public Books (2) | User Books (2) | Admin Books (2) | Total |
|------|-----------------|----------------|-----------------|-------|
| admin | ✅ | ✅ | ✅ | 6 |
| user | ✅ | ✅ | ❌ | 4 |
| guest | ✅ | ❌ | ❌ | 2 |
| other | ✅ | ❌ | ❌ | 2 |

## Technical Details

### Technologies Used

- **Traefik v3.0**: Reverse proxy and load balancer
- **Go 1.21**: Programming language for services
- **Docker & Docker Compose**: Containerization and orchestration
- **Alpine Linux**: Base image for minimal containers

### Key Features Implemented

1. ✅ Forward auth middleware configuration
2. ✅ Token validation and authentication
3. ✅ Request header enrichment
4. ✅ Access logging
5. ✅ Role-based access control
6. ✅ Traefik dashboard
7. ✅ Health checks
8. ✅ Docker containerization
9. ✅ Automated testing
10. ✅ Comprehensive documentation

### How It Works

1. Client sends request to `/books` with Authorization header
2. Traefik intercepts and forwards to auth service at `/auth`
3. Auth service validates token and logs access
4. If valid, auth service returns 200 with enriched headers
5. Traefik copies headers and forwards to book service
6. Book service filters books based on role
7. Response flows back through Traefik to client

## Next Steps

### To Run This POC

1. Clone/navigate to the project directory
2. Run `make up` or `docker-compose up -d`
3. Wait a few seconds for services to start
4. Run `make test` to verify everything works
5. Try the examples in EXAMPLES.md
6. View the Traefik dashboard at http://localhost:8080

### To Extend This POC

1. **Add JWT support**: Replace simple token with JWT validation
2. **Add database**: Store users and tokens in a database
3. **Add rate limiting**: Use Traefik's rate limit middleware
4. **Add HTTPS**: Enable TLS/SSL
5. **Add more services**: Protect additional services with same auth
6. **Add refresh tokens**: Implement token refresh mechanism
7. **Add user management**: Create user registration/login endpoints
8. **Add monitoring**: Integrate Prometheus and Grafana

## Testing Checklist

- [ ] Services start successfully (`make up`)
- [ ] Request without token fails with 401
- [ ] Admin token returns all 6 books
- [ ] User token returns 4 books
- [ ] Guest token returns 2 books
- [ ] Invalid token fails with 401
- [ ] Access logs are recorded
- [ ] Traefik dashboard is accessible
- [ ] Health checks return OK
- [ ] All automated tests pass (`make test`)

## Troubleshooting

### Services won't start
```bash
# Check if ports are in use
lsof -i :80
lsof -i :8080
lsof -i :8081
lsof -i :8082

# Rebuild from scratch
make clean
make build
make up
```

### Authentication always fails
- Check token format: `Bearer <userID>:<role>:secret-token`
- Verify token is exactly `secret-token`
- Check auth service logs: `make logs-auth`

### Books not filtered correctly
- Verify headers are being passed: Check book service logs
- Check Traefik configuration: Review `dynamic-config.yml`
- Ensure services are communicating: Check network connectivity

## Resources

- [Traefik Documentation](https://doc.traefik.io/traefik/)
- [Forward Auth Middleware](https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/forwardauth/)
- [Docker Documentation](https://docs.docker.com/)
- [Go Documentation](https://golang.org/doc/)
