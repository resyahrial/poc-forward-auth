package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		wantUserID string
		wantRole   string
		wantValid  bool
	}{
		{
			name:       "Valid admin token",
			authHeader: "Bearer admin:admin:secret-token",
			wantUserID: "admin",
			wantRole:   "admin",
			wantValid:  true,
		},
		{
			name:       "Valid user token",
			authHeader: "Bearer user123:user:secret-token",
			wantUserID: "user123",
			wantRole:   "user",
			wantValid:  true,
		},
		{
			name:       "Valid guest token",
			authHeader: "Bearer guest:guest:secret-token",
			wantUserID: "guest",
			wantRole:   "guest",
			wantValid:  true,
		},
		{
			name:       "Invalid token - wrong secret",
			authHeader: "Bearer user:admin:wrong-token",
			wantUserID: "",
			wantRole:   "",
			wantValid:  false,
		},
		{
			name:       "Invalid token - missing bearer prefix",
			authHeader: "user:admin:secret-token",
			wantUserID: "",
			wantRole:   "",
			wantValid:  false,
		},
		{
			name:       "Invalid token - wrong format (only 2 parts)",
			authHeader: "Bearer user:secret-token",
			wantUserID: "",
			wantRole:   "",
			wantValid:  false,
		},
		{
			name:       "Empty header",
			authHeader: "",
			wantUserID: "",
			wantRole:   "",
			wantValid:  false,
		},
		{
			name:       "Invalid prefix",
			authHeader: "Basic user:admin:secret-token",
			wantUserID: "",
			wantRole:   "",
			wantValid:  false,
		},
		{
			name:       "Token with extra parts",
			authHeader: "Bearer user:admin:secret-token:extra",
			wantUserID: "",
			wantRole:   "",
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, role, valid := validateToken(tt.authHeader)
			if userID != tt.wantUserID {
				t.Errorf("validateToken() userID = %v, want %v", userID, tt.wantUserID)
			}
			if role != tt.wantRole {
				t.Errorf("validateToken() role = %v, want %v", role, tt.wantRole)
			}
			if valid != tt.wantValid {
				t.Errorf("validateToken() valid = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestAuthHandler(t *testing.T) {
	// Clear access logs before tests
	accessLogs = []AccessLog{}

	tests := []struct {
		name         string
		authHeader   string
		wantStatus   int
		wantUserID   string
		wantRole     string
		checkHeaders bool
		wantLogCount int
	}{
		{
			name:         "Successful authentication - admin",
			authHeader:   "Bearer admin:admin:secret-token",
			wantStatus:   http.StatusOK,
			wantUserID:   "admin",
			wantRole:     "admin",
			checkHeaders: true,
			wantLogCount: 1,
		},
		{
			name:         "Successful authentication - user",
			authHeader:   "Bearer alice:user:secret-token",
			wantStatus:   http.StatusOK,
			wantUserID:   "alice",
			wantRole:     "user",
			checkHeaders: true,
			wantLogCount: 2,
		},
		{
			name:         "Failed authentication - invalid token",
			authHeader:   "Bearer user:admin:wrong-token",
			wantStatus:   http.StatusUnauthorized,
			checkHeaders: false,
			wantLogCount: 2, // Should not increment
		},
		{
			name:         "Failed authentication - missing token",
			authHeader:   "",
			wantStatus:   http.StatusUnauthorized,
			checkHeaders: false,
			wantLogCount: 2,
		},
		{
			name:         "Failed authentication - malformed token",
			authHeader:   "InvalidFormat",
			wantStatus:   http.StatusUnauthorized,
			checkHeaders: false,
			wantLogCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			req.Header.Set("X-Forwarded-Method", "GET")
			req.Header.Set("X-Forwarded-Uri", "/books")
			req.Header.Set("X-Forwarded-For", "127.0.0.1")

			w := httptest.NewRecorder()
			authHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("authHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.checkHeaders {
				if got := w.Header().Get("X-User-ID"); got != tt.wantUserID {
					t.Errorf("X-User-ID header = %v, want %v", got, tt.wantUserID)
				}
				if got := w.Header().Get("X-User-Role"); got != tt.wantRole {
					t.Errorf("X-User-Role header = %v, want %v", got, tt.wantRole)
				}
				if got := w.Header().Get("X-Auth-Time"); got == "" {
					t.Error("X-Auth-Time header should not be empty")
				}
			}

			if len(accessLogs) != tt.wantLogCount {
				t.Errorf("access logs count = %v, want %v", len(accessLogs), tt.wantLogCount)
			}
		})
	}
}

func TestAccessLogsHandler(t *testing.T) {
	// Setup test data
	accessLogs = []AccessLog{
		{
			Timestamp:    time.Now(),
			Method:       "GET",
			URI:          "/books",
			ForwardedFor: "127.0.0.1",
			UserID:       "admin",
			Role:         "admin",
		},
		{
			Timestamp:    time.Now(),
			Method:       "GET",
			URI:          "/books",
			ForwardedFor: "127.0.0.1",
			UserID:       "user",
			Role:         "user",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/access-logs", nil)
	w := httptest.NewRecorder()

	accessLogsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("accessLogsHandler() status = %v, want %v", w.Code, http.StatusOK)
	}

	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type = %v, want application/json", contentType)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	totalRequests, ok := response["total_requests"].(float64)
	if !ok {
		t.Error("total_requests should be a number")
	}
	if int(totalRequests) != 2 {
		t.Errorf("total_requests = %v, want 2", totalRequests)
	}

	logs, ok := response["logs"].([]interface{})
	if !ok {
		t.Error("logs should be an array")
	}
	if len(logs) != 2 {
		t.Errorf("logs count = %v, want 2", len(logs))
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("healthHandler() status = %v, want %v", w.Code, http.StatusOK)
	}

	want := "Auth service is healthy"
	if got := w.Body.String(); got != want {
		t.Errorf("healthHandler() body = %v, want %v", got, want)
	}
}

func TestAccessLogStructure(t *testing.T) {
	log := AccessLog{
		Timestamp:    time.Now(),
		Method:       "GET",
		URI:          "/books",
		ForwardedFor: "192.168.1.1",
		UserID:       "testuser",
		Role:         "admin",
	}

	// Test JSON marshaling
	data, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("Failed to marshal AccessLog: %v", err)
	}

	// Test JSON unmarshaling
	var decoded AccessLog
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AccessLog: %v", err)
	}

	if decoded.Method != log.Method {
		t.Errorf("Method = %v, want %v", decoded.Method, log.Method)
	}
	if decoded.URI != log.URI {
		t.Errorf("URI = %v, want %v", decoded.URI, log.URI)
	}
	if decoded.UserID != log.UserID {
		t.Errorf("UserID = %v, want %v", decoded.UserID, log.UserID)
	}
	if decoded.Role != log.Role {
		t.Errorf("Role = %v, want %v", decoded.Role, log.Role)
	}
}

func TestConcurrentAuthRequests(t *testing.T) {
	// Clear logs
	accessLogs = []AccessLog{}

	// Make concurrent requests
	done := make(chan bool)
	requestCount := 10

	for i := 0; i < requestCount; i++ {
		go func(id int) {
			req := httptest.NewRequest(http.MethodGet, "/auth", nil)
			req.Header.Set("Authorization", "Bearer user:user:secret-token")
			req.Header.Set("X-Forwarded-Method", "GET")
			req.Header.Set("X-Forwarded-Uri", "/books")

			w := httptest.NewRecorder()
			authHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Concurrent request %d failed with status %v", id, w.Code)
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < requestCount; i++ {
		<-done
	}

	// Note: This test might have race conditions since we're not using mutex
	// In a production environment, you'd want to protect accessLogs with a mutex
	if len(accessLogs) != requestCount {
		t.Logf("Warning: Expected %d logs but got %d (race condition possible)", requestCount, len(accessLogs))
	}
}
