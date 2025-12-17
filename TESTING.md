# Test Documentation

This document describes the test suite for the Traefik Forward Auth POC.

## Test Overview

The project includes two types of tests:
1. **Unit Tests** - Test individual functions and handlers
2. **Integration Tests** - Test the complete system with Docker

## Test Coverage

### Auth Service Tests
- **Coverage**: 72.3%
- **Test File**: `auth-service/main_test.go`
- **Total Tests**: 6 test functions with 23 sub-tests

### Book Service Tests
- **Coverage**: 59.4%
- **Test File**: `book-service/main_test.go`
- **Total Tests**: 10 test functions with 15 sub-tests

## Running Tests

### All Tests (Unit + Integration)
```bash
make test-all
```

### Unit Tests Only
```bash
make test-unit
# or
./run-tests.sh
```

### Integration Tests Only
```bash
make test
# or
./test.sh
```

### Individual Service Tests

**Auth Service Only:**
```bash
make test-auth
# or
cd auth-service && go test -v -cover
```

**Book Service Only:**
```bash
make test-book
# or
cd book-service && go test -v -cover
```

## Auth Service Tests

### 1. TestValidateToken
Tests the token validation function with various scenarios:

✅ **Valid Tokens:**
- Admin token: `Bearer admin:admin:secret-token`
- User token: `Bearer user123:user:secret-token`
- Guest token: `Bearer guest:guest:secret-token`

❌ **Invalid Tokens:**
- Wrong secret: `Bearer user:admin:wrong-token`
- Missing Bearer prefix: `user:admin:secret-token`
- Wrong format (2 parts): `Bearer user:secret-token`
- Empty header: `""`
- Invalid prefix: `Basic user:admin:secret-token`
- Extra parts: `Bearer user:admin:secret-token:extra`

### 2. TestAuthHandler
Tests the main authentication handler:

✅ **Successful Authentication:**
- Admin authentication with proper headers
- User authentication with proper headers
- Verifies X-User-ID, X-User-Role, X-Auth-Time headers are set
- Verifies access logging

❌ **Failed Authentication:**
- Invalid token returns 401
- Missing token returns 401
- Malformed token returns 401
- Verifies no access log entry for failed attempts

### 3. TestAccessLogsHandler
Tests the access logs endpoint:
- Verifies JSON response format
- Checks `total_requests` count
- Validates `logs` array structure
- Ensures Content-Type is application/json

### 4. TestHealthHandler
Tests health check endpoint:
- Verifies 200 OK status
- Checks response body: "Auth service is healthy"

### 5. TestAccessLogStructure
Tests AccessLog struct:
- JSON marshaling
- JSON unmarshaling
- Field preservation

### 6. TestConcurrentAuthRequests
Tests concurrent request handling:
- Makes 10 concurrent authentication requests
- Verifies all requests succeed
- Notes: May have race conditions (documented in test)

## Book Service Tests

### 1. TestFilterBooksByRole
Tests role-based book filtering:

**Admin Role:**
- Sees all 6 books (public + user + admin)

**User Role:**
- Sees 4 books (public + user)

**Guest Role:**
- Sees 2 books (public only)

**Unknown/Empty Role:**
- Sees 2 books (public only)

### 2. TestFilterBooksByRole_AdminAccess
Detailed admin access test:
- Verifies all 6 books are returned
- Checks distribution: 2 public, 2 user, 2 admin

### 3. TestFilterBooksByRole_UserAccess
Detailed user access test:
- Verifies 4 books returned
- Ensures no admin books included

### 4. TestFilterBooksByRole_GuestAccess
Detailed guest access test:
- Verifies 2 books returned
- Ensures all books are public

### 5. TestBooksHandler
Tests the main books handler:

**With All Headers:**
- Admin request → 6 books
- User request → 4 books
- Guest request → 2 books

**Without Headers (Direct Access):**
- Returns 2 books (public only)

**Verifies Response Structure:**
- user_id field
- role field
- auth_time field
- total_books field
- books array
- access_level field

### 6. TestHealthHandler
Tests health check:
- Verifies 200 OK status
- Checks response: "Book service is healthy"

### 7. TestBookStructure
Tests Book struct:
- JSON marshaling
- JSON unmarshaling
- Field preservation

### 8. TestBooksData
Validates books data initialization:
- Total count: 6 books
- Distribution: 2 public, 2 user, 2 admin
- All required fields present

### 9. TestBooksHandler_ResponseStructure
Tests response format:
- Verifies all expected fields present
- Validates field names

### 10. TestBooksHandler_DifferentMethods
Tests handler with different HTTP methods:
- GET
- POST
- PUT
- DELETE
- All should return 200 OK

### 11. TestFilterBooksByRole_Consistency
Tests filter consistency:
- Runs filter 100 times
- Ensures same results each time

## Test Results

### Latest Test Run

```
Auth Service:
  ✓ 6 test functions
  ✓ 23 sub-tests
  ✓ 72.3% code coverage
  ✓ All tests passed

Book Service:
  ✓ 10 test functions
  ✓ 15 sub-tests
  ✓ 59.4% code coverage
  ✓ All tests passed
```

## Integration Tests

The integration test script ([test.sh](test.sh)) performs end-to-end testing:

### Test Scenarios

1. **No Authentication**
   - Request: No Authorization header
   - Expected: 401 Unauthorized

2. **Invalid Token**
   - Request: `Bearer user:admin:wrong-token`
   - Expected: 401 Unauthorized

3. **Admin Access**
   - Request: `Bearer admin:admin:secret-token`
   - Expected: 200 OK, 6 books

4. **User Access**
   - Request: `Bearer alice:user:secret-token`
   - Expected: 200 OK, 4 books

5. **Guest Access**
   - Request: `Bearer bob:guest:secret-token`
   - Expected: 200 OK, 2 books

6. **Access Logs**
   - Request: GET /access-logs
   - Expected: List of authentication attempts

## Test Data

### Books
```go
{ID: "1", Title: "Public Book 1", Author: "Author A", Access: "public"}
{ID: "2", Title: "Public Book 2", Author: "Author B", Access: "public"}
{ID: "3", Title: "User Book 1", Author: "Author C", Access: "user"}
{ID: "4", Title: "User Book 2", Author: "Author D", Access: "user"}
{ID: "5", Title: "Admin Book 1", Author: "Author E", Access: "admin"}
{ID: "6", Title: "Admin Book 2", Author: "Author F", Access: "admin"}
```

### Valid Token
```
secret-token
```

## Testing Best Practices

### Unit Testing
1. Test both success and failure cases
2. Test edge cases (empty strings, nil values)
3. Test with various input combinations
4. Use table-driven tests for multiple scenarios
5. Mock external dependencies

### Integration Testing
1. Test complete workflows
2. Use realistic test data
3. Verify HTTP status codes
4. Check response headers and body
5. Test authentication flow end-to-end

## Continuous Integration

To integrate with CI/CD:

```bash
# Run in CI pipeline
./run-tests.sh || exit 1
```

Or use Go's native test command:

```bash
# Auth service
cd auth-service && go test -v -cover ./... || exit 1

# Book service
cd book-service && go test -v -cover ./... || exit 1
```

## Coverage Goals

Current coverage:
- Auth Service: 72.3% ✅
- Book Service: 59.4% 🎯

Target coverage: 80%

### To Increase Coverage

**Auth Service:**
- Add tests for main() function initialization
- Test concurrent access to accessLogs slice with mutex
- Add benchmark tests

**Book Service:**
- Add tests for main() function initialization
- Test error handling scenarios
- Add benchmark tests for filterBooksByRole

## Known Issues

1. **Race Condition in TestConcurrentAuthRequests**
   - The `accessLogs` slice is not protected by mutex
   - Concurrent writes may cause inconsistent counts
   - Logged as a warning in tests
   - Solution: Add sync.Mutex to protect accessLogs

## Adding New Tests

### Template for New Test

```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "expected",
            wantErr: false,
        },
        // Add more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := newFeature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("newFeature() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("newFeature() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Troubleshooting Tests

### Tests Fail to Run
```bash
# Ensure you're in the correct directory
cd auth-service
go test -v

# or
cd book-service
go test -v
```

### Tests Pass Locally But Fail in CI
- Check Go version (should be 1.23)
- Verify dependencies are available
- Check for race conditions with `-race` flag:
  ```bash
  go test -race -v
  ```

### Coverage Report
```bash
# Generate HTML coverage report
cd auth-service
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

## Performance Testing

### Benchmarks

To add benchmarks:

```go
func BenchmarkValidateToken(b *testing.B) {
    for i := 0; i < b.N; i++ {
        validateToken("Bearer user:admin:secret-token")
    }
}
```

Run benchmarks:
```bash
cd auth-service
go test -bench=. -benchmem
```

## Future Improvements

1. Add integration tests with actual Docker containers
2. Add load testing with k6 or similar tools
3. Add security testing (penetration tests)
4. Add mutation testing
5. Increase code coverage to 80%+
6. Add race detection to CI/CD
7. Add fuzzing tests
8. Add contract tests
9. Add end-to-end tests with real Traefik
10. Add performance benchmarks
