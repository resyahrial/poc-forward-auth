package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type AccessLog struct {
	Timestamp    time.Time `json:"timestamp"`
	Method       string    `json:"method"`
	URI          string    `json:"uri"`
	ForwardedFor string    `json:"forwarded_for"`
	UserID       string    `json:"user_id"`
	Role         string    `json:"role"`
}

// In-memory storage for access logs
var accessLogs []AccessLog

// Simple token validation
// Format: "Bearer user:role:token"
// Example: "Bearer user123:admin:secret-token"
func validateToken(authHeader string) (userID, role string, valid bool) {
	if authHeader == "" {
		return "", "", false
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", "", false
	}

	tokenParts := strings.Split(parts[1], ":")
	if len(tokenParts) != 3 {
		return "", "", false
	}

	userID = tokenParts[0]
	role = tokenParts[1]
	token := tokenParts[2]

	// Simple validation - in real world, this would validate against a database or JWT
	if token == "secret-token" {
		return userID, role, true
	}

	return "", "", false
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Auth request: Method=%s, URI=%s, Headers=%v", r.Method, r.RequestURI, r.Header)

	// Get authorization header
	authHeader := r.Header.Get("Authorization")

	// Validate token
	userID, role, valid := validateToken(authHeader)

	if !valid {
		log.Printf("Authentication failed: invalid token")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized: Invalid or missing token"))
		return
	}

	// Log the access
	accessLog := AccessLog{
		Timestamp:    time.Now(),
		Method:       r.Header.Get("X-Forwarded-Method"),
		URI:          r.Header.Get("X-Forwarded-Uri"),
		ForwardedFor: r.Header.Get("X-Forwarded-For"),
		UserID:       userID,
		Role:         role,
	}
	accessLogs = append(accessLogs, accessLog)

	log.Printf("Authentication successful: UserID=%s, Role=%s", userID, role)

	// Enrich request headers
	// These headers will be forwarded to the backend service
	w.Header().Set("X-User-ID", userID)
	w.Header().Set("X-User-Role", role)
	w.Header().Set("X-Auth-Time", time.Now().Format(time.RFC3339))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authorized"))
}

func accessLogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_requests": len(accessLogs),
		"logs":           accessLogs,
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Auth service is healthy"))
}

func main() {
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/access-logs", accessLogsHandler)
	http.HandleFunc("/health", healthHandler)

	port := "8081"
	log.Printf("Auth service starting on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  - /auth (ForwardAuth endpoint)")
	log.Printf("  - /access-logs (View access logs)")
	log.Printf("  - /health (Health check)")
	log.Printf("\nToken format: Bearer <userID>:<role>:secret-token")
	log.Printf("Example: Authorization: Bearer user123:admin:secret-token")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
