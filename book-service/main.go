package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Access string `json:"access"` // public, user, admin
}

var books = []Book{
	{ID: "1", Title: "Public Book 1", Author: "Author A", Access: "public"},
	{ID: "2", Title: "Public Book 2", Author: "Author B", Access: "public"},
	{ID: "3", Title: "User Book 1", Author: "Author C", Access: "user"},
	{ID: "4", Title: "User Book 2", Author: "Author D", Access: "user"},
	{ID: "5", Title: "Admin Book 1", Author: "Author E", Access: "admin"},
	{ID: "6", Title: "Admin Book 2", Author: "Author F", Access: "admin"},
}

func filterBooksByRole(role string) []Book {
	var filteredBooks []Book

	for _, book := range books {
		switch role {
		case "admin":
			// Admin can see all books
			filteredBooks = append(filteredBooks, book)
		case "user":
			// User can see public and user books
			if book.Access == "public" || book.Access == "user" {
				filteredBooks = append(filteredBooks, book)
			}
		default:
			// Unknown roles can only see public books
			if book.Access == "public" {
				filteredBooks = append(filteredBooks, book)
			}
		}
	}

	return filteredBooks
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	// Get enriched headers from auth service
	userID := r.Header.Get("X-User-ID")
	role := r.Header.Get("X-User-Role")
	authTime := r.Header.Get("X-Auth-Time")

	log.Printf("Request from UserID=%s, Role=%s, AuthTime=%s", userID, role, authTime)

	// Filter books based on user role
	filteredBooks := filterBooksByRole(role)

	response := map[string]interface{}{
		"user_id":      userID,
		"role":         role,
		"auth_time":    authTime,
		"total_books":  len(filteredBooks),
		"books":        filteredBooks,
		"access_level": role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Book service is healthy"))
}

func main() {
	http.HandleFunc("/books", booksHandler)
	http.HandleFunc("/health", healthHandler)

	port := "8082"
	log.Printf("Book service starting on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  - /books (Get books based on role)")
	log.Printf("  - /health (Health check)")
	log.Printf("\nAccess levels:")
	log.Printf("  - public: Everyone can see")
	log.Printf("  - user: Users and admins can see")
	log.Printf("  - admin: Only admins can see")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
