# Quick Start Guide

## 1. Start the POC

```bash
docker-compose up --build
```

Wait for all services to start (you should see logs from traefik, auth-service, and book-service).

## 2. Run the Test Script

In a new terminal:

```bash
./test.sh
```

This will run automated tests covering all scenarios.

## 3. Manual Testing

### Admin Access (sees all books)
```bash
curl http://localhost/books \
  -H "Authorization: Bearer admin:admin:secret-token" | jq
```

### User Access (sees public and user books)
```bash
curl http://localhost/books \
  -H "Authorization: Bearer user123:user:secret-token" | jq
```

### Guest Access (sees public books only)
```bash
curl http://localhost/books \
  -H "Authorization: Bearer guest:guest:secret-token" | jq
```

### View Access Logs
```bash
curl http://localhost/access-logs | jq
```

## 4. Explore Traefik Dashboard

Open in browser: http://localhost:8080

## 5. Stop the POC

```bash
docker-compose down
```

## Token Format

```
Authorization: Bearer <userID>:<role>:<token>
```

- Valid token: `secret-token`
- Roles: `admin`, `user`, `guest`, or any other value
- UserID: any string

## Access Matrix

| Role   | Public Books | User Books | Admin Books |
|--------|--------------|------------|-------------|
| admin  | ✅           | ✅         | ✅          |
| user   | ✅           | ✅         | ❌          |
| guest  | ✅           | ❌         | ❌          |
