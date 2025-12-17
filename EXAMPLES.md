# Example API Calls

This file contains example curl commands for testing the Traefik Forward Auth POC.

## Prerequisites

Make sure the services are running:
```bash
docker-compose up -d
# or
make up
```

---

## 1. Authentication Tests

### ✅ Valid Admin Token
```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token"
```
**Expected**: 200 OK, returns all 6 books

---

### ✅ Valid User Token
```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer alice:user:secret-token"
```
**Expected**: 200 OK, returns 4 books (public + user)

---

### ✅ Valid Guest Token
```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer bob:guest:secret-token"
```
**Expected**: 200 OK, returns 2 books (public only)

---

### ❌ Invalid Token
```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer user:admin:wrong-token"
```
**Expected**: 401 Unauthorized

---

### ❌ Missing Token
```bash
curl -v http://localhost/books
```
**Expected**: 401 Unauthorized

---

### ❌ Malformed Token
```bash
curl -v http://localhost/books \
  -H "Authorization: InvalidFormat"
```
**Expected**: 401 Unauthorized

---

## 2. Different User Scenarios

### Scenario A: Admin accessing books
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

---

### Scenario B: Regular user accessing books
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

---

### Scenario C: Guest accessing books
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

---

## 3. Access Logging

### View all access logs
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

---

### Count total requests
```bash
curl -s http://localhost/access-logs | jq '.total_requests'
```

---

### Filter logs by user
```bash
curl -s http://localhost/access-logs | jq '.logs[] | select(.user_id == "admin")'
```

---

## 4. Health Checks

### Auth Service Health
```bash
curl http://localhost:8081/health
```
**Expected**: "Auth service is healthy"

---

### Book Service Health
```bash
curl http://localhost:8082/health
```
**Expected**: "Book service is healthy"

---

## 5. Direct Service Access (Bypassing Traefik)

⚠️ **Note**: This bypasses authentication and should not be allowed in production

### Direct access to Auth Service
```bash
# Health check
curl http://localhost:8081/health

# Access logs
curl http://localhost:8081/access-logs | jq
```

---

### Direct access to Book Service (No Auth!)
```bash
curl http://localhost:8082/books | jq
```
**Note**: This bypasses Traefik and authentication, so no headers are enriched.
You'll see empty values for user_id, role, etc.

---

## 6. Response Analysis

### Get only book count
```bash
curl -s http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token" | jq '.total_books'
```

---

### Get only book titles
```bash
curl -s http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token" | jq '.books[].title'
```

---

### Get books by access level
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

---

## 7. Header Inspection

### View enriched headers (requires verbose mode)
```bash
curl -v http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token" 2>&1 | grep -i "X-"
```

This won't show the enriched headers in the response, but you can see them in the auth service logs:
```bash
docker-compose logs auth-service | tail -20
```

---

## 8. Performance Testing

### Sequential requests
```bash
for i in {1..10}; do
  curl -s http://localhost/books \
    -H "Authorization: Bearer user$i:user:secret-token" | jq '.user_id'
done
```

---

### Check access log count after multiple requests
```bash
# Make 5 requests
for i in {1..5}; do
  curl -s http://localhost/books \
    -H "Authorization: Bearer user$i:admin:secret-token" > /dev/null
done

# Check total count
curl -s http://localhost/access-logs | jq '.total_requests'
```

---

## 9. Error Scenarios

### Empty Authorization header
```bash
curl -v http://localhost/books -H "Authorization: "
```
**Expected**: 401 Unauthorized

---

### Bearer without token
```bash
curl -v http://localhost/books -H "Authorization: Bearer "
```
**Expected**: 401 Unauthorized

---

### Token with only 2 parts
```bash
curl -v http://localhost/books -H "Authorization: Bearer user:secret-token"
```
**Expected**: 401 Unauthorized (needs 3 parts: userID:role:token)

---

## 10. Comparison Tests

### Compare book count across roles
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

---

## 11. Traefik Dashboard

### Open in browser
```
http://localhost:8080
```

### API endpoints (if enabled)
```bash
# Get routers
curl http://localhost:8080/api/http/routers

# Get middlewares
curl http://localhost:8080/api/http/middlewares

# Get services
curl http://localhost:8080/api/http/services
```

---

## 12. Docker Logs

### View all logs
```bash
docker-compose logs -f
```

---

### View auth service logs only
```bash
docker-compose logs -f auth-service
```

---

### View book service logs only
```bash
docker-compose logs -f book-service
```

---

### View Traefik logs only
```bash
docker-compose logs -f traefik
```

---

## Tips

1. **Use jq for pretty JSON**: Install jq (`brew install jq` on Mac) for better JSON formatting
2. **Use -v for debugging**: Add `-v` flag to see full request/response headers
3. **Check logs**: Always check service logs when debugging issues
4. **Token format**: Remember the format is `Bearer <userID>:<role>:<token>`
5. **Valid token**: The only valid token is `secret-token`

---

## Quick Test Sequence

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
