# Quick Test Reference

## Running Tests

| Command | Description |
|---------|-------------|
| `make test-unit` | Run all unit tests |
| `make test-auth` | Test auth service only |
| `make test-book` | Test book service only |
| `make test` | Run integration tests |
| `make test-all` | Run all tests |
| `./run-tests.sh` | Run unit tests (script) |
| `./test.sh` | Run integration tests (script) |

## Test Files

| File | Description | Coverage |
|------|-------------|----------|
| `auth-service/main_test.go` | Auth service unit tests | 72.3% |
| `book-service/main_test.go` | Book service unit tests | 59.4% |
| `run-tests.sh` | Unit test runner script | - |
| `test.sh` | Integration test script | - |

## Test Statistics

### Auth Service
- **6** test functions
- **23** sub-tests
- **72.3%** code coverage
- **0** failures

### Book Service  
- **10** test functions
- **15+** sub-tests
- **59.4%** code coverage
- **0** failures

## Quick Commands

```bash
# Run all unit tests
make test-unit

# Run specific service tests
cd auth-service && go test -v
cd book-service && go test -v

# Run with coverage report
cd auth-service && go test -cover
cd book-service && go test -cover

# Generate HTML coverage
cd auth-service
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Scenarios

### Auth Service Tests
✅ Valid admin token
✅ Valid user token  
✅ Valid guest token
✅ Invalid token (wrong secret)
✅ Missing token
✅ Malformed token
✅ Header enrichment
✅ Access logging
✅ Concurrent requests

### Book Service Tests
✅ Admin sees 6 books
✅ User sees 4 books
✅ Guest sees 2 books
✅ Unknown role sees 2 books
✅ Response structure
✅ Different HTTP methods
✅ Filter consistency

## Expected Results

```
All tests passed!
Auth Service: 72.3% coverage
Book Service: 59.4% coverage
```

## Documentation

- [TESTING.md](TESTING.md) - Full testing documentation
- [TEST-SUMMARY.md](TEST-SUMMARY.md) - Test implementation summary
- [README.md](README.md) - Main project documentation
