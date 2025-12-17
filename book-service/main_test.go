package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFilterBooksByRole(t *testing.T) {
	tests := []struct {
		name      string
		role      string
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "Admin sees all books",
			role:      "admin",
			wantCount: 6,
			wantIDs:   []string{"1", "2", "3", "4", "5", "6"},
		},
		{
			name:      "User sees public and user books",
			role:      "user",
			wantCount: 4,
			wantIDs:   []string{"1", "2", "3", "4"},
		},
		{
			name:      "Guest sees only public books",
			role:      "guest",
			wantCount: 2,
			wantIDs:   []string{"1", "2"},
		},
		{
			name:      "Unknown role sees only public books",
			role:      "unknown",
			wantCount: 2,
			wantIDs:   []string{"1", "2"},
		},
		{
			name:      "Empty role sees only public books",
			role:      "",
			wantCount: 2,
			wantIDs:   []string{"1", "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterBooksByRole(tt.role)

			if len(result) != tt.wantCount {
				t.Errorf("filterBooksByRole(%s) returned %d books, want %d", tt.role, len(result), tt.wantCount)
			}

			// Check if returned book IDs match expected
			for i, expectedID := range tt.wantIDs {
				if i >= len(result) {
					t.Errorf("Missing book at index %d", i)
					continue
				}
				if result[i].ID != expectedID {
					t.Errorf("Book at index %d has ID %s, want %s", i, result[i].ID, expectedID)
				}
			}
		})
	}
}

func TestFilterBooksByRole_AdminAccess(t *testing.T) {
	result := filterBooksByRole("admin")

	// Admin should see all books
	if len(result) != 6 {
		t.Errorf("Admin should see 6 books, got %d", len(result))
	}

	// Check all access levels are present
	accessLevels := make(map[string]int)
	for _, book := range result {
		accessLevels[book.Access]++
	}

	if accessLevels["public"] != 2 {
		t.Errorf("Expected 2 public books, got %d", accessLevels["public"])
	}
	if accessLevels["user"] != 2 {
		t.Errorf("Expected 2 user books, got %d", accessLevels["user"])
	}
	if accessLevels["admin"] != 2 {
		t.Errorf("Expected 2 admin books, got %d", accessLevels["admin"])
	}
}

func TestFilterBooksByRole_UserAccess(t *testing.T) {
	result := filterBooksByRole("user")

	// User should see public and user books only
	if len(result) != 4 {
		t.Errorf("User should see 4 books, got %d", len(result))
	}

	// Check no admin books are included
	for _, book := range result {
		if book.Access == "admin" {
			t.Errorf("User should not see admin books, but got book: %s", book.Title)
		}
	}
}

func TestFilterBooksByRole_GuestAccess(t *testing.T) {
	result := filterBooksByRole("guest")

	// Guest should see only public books
	if len(result) != 2 {
		t.Errorf("Guest should see 2 books, got %d", len(result))
	}

	// Check all books are public
	for _, book := range result {
		if book.Access != "public" {
			t.Errorf("Guest should only see public books, but got book with access: %s", book.Access)
		}
	}
}

func TestBooksHandler(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		role          string
		authTime      string
		wantStatus    int
		wantBookCount int
		checkResponse bool
	}{
		{
			name:          "Admin request with all headers",
			userID:        "admin",
			role:          "admin",
			authTime:      "2025-12-17T10:00:00Z",
			wantStatus:    http.StatusOK,
			wantBookCount: 6,
			checkResponse: true,
		},
		{
			name:          "User request with all headers",
			userID:        "alice",
			role:          "user",
			authTime:      "2025-12-17T10:01:00Z",
			wantStatus:    http.StatusOK,
			wantBookCount: 4,
			checkResponse: true,
		},
		{
			name:          "Guest request with all headers",
			userID:        "bob",
			role:          "guest",
			authTime:      "2025-12-17T10:02:00Z",
			wantStatus:    http.StatusOK,
			wantBookCount: 2,
			checkResponse: true,
		},
		{
			name:          "Request without headers (direct access)",
			userID:        "",
			role:          "",
			authTime:      "",
			wantStatus:    http.StatusOK,
			wantBookCount: 2, // Default to public books
			checkResponse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/books", nil)
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}
			if tt.role != "" {
				req.Header.Set("X-User-Role", tt.role)
			}
			if tt.authTime != "" {
				req.Header.Set("X-Auth-Time", tt.authTime)
			}

			w := httptest.NewRecorder()
			booksHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("booksHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.checkResponse {
				if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
					t.Errorf("Content-Type = %v, want application/json", contentType)
				}

				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Check user_id
				if got := response["user_id"]; got != tt.userID {
					t.Errorf("user_id = %v, want %v", got, tt.userID)
				}

				// Check role
				if got := response["role"]; got != tt.role {
					t.Errorf("role = %v, want %v", got, tt.role)
				}

				// Check total_books
				totalBooks, ok := response["total_books"].(float64)
				if !ok {
					t.Error("total_books should be a number")
				}
				if int(totalBooks) != tt.wantBookCount {
					t.Errorf("total_books = %v, want %v", totalBooks, tt.wantBookCount)
				}

				// Check books array
				booksArray, ok := response["books"].([]interface{})
				if !ok {
					t.Error("books should be an array")
				}
				if len(booksArray) != tt.wantBookCount {
					t.Errorf("books count = %v, want %v", len(booksArray), tt.wantBookCount)
				}

				// Check access_level
				if got := response["access_level"]; got != tt.role {
					t.Errorf("access_level = %v, want %v", got, tt.role)
				}
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("healthHandler() status = %v, want %v", w.Code, http.StatusOK)
	}

	want := "Book service is healthy"
	if got := w.Body.String(); got != want {
		t.Errorf("healthHandler() body = %v, want %v", got, want)
	}
}

func TestBookStructure(t *testing.T) {
	book := Book{
		ID:     "1",
		Title:  "Test Book",
		Author: "Test Author",
		Access: "public",
	}

	// Test JSON marshaling
	data, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("Failed to marshal Book: %v", err)
	}

	// Test JSON unmarshaling
	var decoded Book
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Book: %v", err)
	}

	if decoded.ID != book.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, book.ID)
	}
	if decoded.Title != book.Title {
		t.Errorf("Title = %v, want %v", decoded.Title, book.Title)
	}
	if decoded.Author != book.Author {
		t.Errorf("Author = %v, want %v", decoded.Author, book.Author)
	}
	if decoded.Access != book.Access {
		t.Errorf("Access = %v, want %v", decoded.Access, book.Access)
	}
}

func TestBooksData(t *testing.T) {
	// Verify books data is correctly initialized
	if len(books) != 6 {
		t.Errorf("Expected 6 books, got %d", len(books))
	}

	// Count books by access level
	accessCount := make(map[string]int)
	for _, book := range books {
		accessCount[book.Access]++
	}

	if accessCount["public"] != 2 {
		t.Errorf("Expected 2 public books, got %d", accessCount["public"])
	}
	if accessCount["user"] != 2 {
		t.Errorf("Expected 2 user books, got %d", accessCount["user"])
	}
	if accessCount["admin"] != 2 {
		t.Errorf("Expected 2 admin books, got %d", accessCount["admin"])
	}

	// Verify all books have required fields
	for i, book := range books {
		if book.ID == "" {
			t.Errorf("Book at index %d has empty ID", i)
		}
		if book.Title == "" {
			t.Errorf("Book at index %d has empty Title", i)
		}
		if book.Author == "" {
			t.Errorf("Book at index %d has empty Author", i)
		}
		if book.Access == "" {
			t.Errorf("Book at index %d has empty Access", i)
		}
	}
}

func TestBooksHandler_ResponseStructure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	req.Header.Set("X-User-ID", "testuser")
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-Auth-Time", "2025-12-17T10:00:00Z")

	w := httptest.NewRecorder()
	booksHandler(w, req)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check all expected fields are present
	expectedFields := []string{"user_id", "role", "auth_time", "total_books", "books", "access_level"}
	for _, field := range expectedFields {
		if _, ok := response[field]; !ok {
			t.Errorf("Response missing field: %s", field)
		}
	}
}

func TestBooksHandler_DifferentMethods(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/books", nil)
			req.Header.Set("X-User-ID", "testuser")
			req.Header.Set("X-User-Role", "user")

			w := httptest.NewRecorder()
			booksHandler(w, req)

			// Handler should respond to all methods
			if w.Code != http.StatusOK {
				t.Errorf("Method %s returned status %v, want %v", method, w.Code, http.StatusOK)
			}
		})
	}
}

func TestFilterBooksByRole_Consistency(t *testing.T) {
	// Run the same filter multiple times to ensure consistency
	role := "user"
	iterations := 100

	var firstResult []Book
	for i := 0; i < iterations; i++ {
		result := filterBooksByRole(role)

		if i == 0 {
			firstResult = result
			continue
		}

		// Check consistency with first result
		if len(result) != len(firstResult) {
			t.Errorf("Iteration %d: got %d books, want %d", i, len(result), len(firstResult))
		}

		for j := range result {
			if result[j].ID != firstResult[j].ID {
				t.Errorf("Iteration %d: book %d ID mismatch: got %s, want %s", i, j, result[j].ID, firstResult[j].ID)
			}
		}
	}
}
