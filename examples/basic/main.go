package main

import (
	"encoding/json"
	"log"
	"net/http"

	openswag "github.com/gopackx/open-swag-go"
	"github.com/gopackx/open-swag-go/pkg/spec"
)

// DTO types
type CreateUserRequest struct {
	Name  string `json:"name" example:"John Doe"`
	Email string `json:"email" example:"john@example.com"`
	Age   int    `json:"age" example:"25"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Age       int    `json:"age"`
	CreatedAt string `json:"created_at"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Handlers
func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UserResponse{
		ID:    "123",
		Name:  "John Doe",
		Email: "john@example.com",
	})
}

func getUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserResponse{
		ID:    "123",
		Name:  "John Doe",
		Email: "john@example.com",
	})
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]UserResponse{})
}

// Endpoint definitions
var CreateUserDoc = openswag.Endpoint{
	Method:      "POST",
	Path:        "/users",
	Summary:     "Create a new user",
	Description: "Create a new user account with the provided information",
	Tags:        []string{"Users"},
	RequestBody: &openswag.RequestBody{
		Description: "User data",
		Required:    true,
		Schema:      CreateUserRequest{},
	},
	Responses: map[int]openswag.Response{
		201: {Description: "User created successfully", Schema: UserResponse{}},
		400: {Description: "Invalid request", Schema: ErrorResponse{}},
	},
	Security: []string{"bearerAuth"},
}

var GetUserDoc = openswag.Endpoint{
	Method:      "GET",
	Path:        "/users/{id}",
	Summary:     "Get user by ID",
	Description: "Retrieve a user by their unique identifier",
	Tags:        []string{"Users"},
	Parameters: []openswag.Parameter{
		{Name: "id", In: "path", Description: "User ID", Required: true, Schema: spec.NewSchema("string")},
	},
	Responses: map[int]openswag.Response{
		200: {Description: "User found", Schema: UserResponse{}},
		404: {Description: "User not found", Schema: ErrorResponse{}},
	},
	Security: []string{"bearerAuth"},
}

var ListUsersDoc = openswag.Endpoint{
	Method:      "GET",
	Path:        "/users",
	Summary:     "List all users",
	Description: "Get a paginated list of users",
	Tags:        []string{"Users"},
	Parameters: []openswag.Parameter{
		{Name: "page", In: "query", Description: "Page number"},
		{Name: "limit", In: "query", Description: "Items per page"},
	},
	Responses: map[int]openswag.Response{
		200: {Description: "Users retrieved", Schema: []UserResponse{}},
	},
	Security: []string{"bearerAuth"},
}

func main() {
	// Create documentation
	docs := openswag.New(openswag.Config{
		Info: openswag.Info{
			Title:       "My API",
			Version:     "1.0.0",
			Description: "A sample API demonstrating open-swag-go",
			Contact: &openswag.Contact{
				Name:  "API Support",
				Email: "support@example.com",
			},
		},
		Servers: []openswag.Server{
			{URL: "http://localhost:8080", Description: "Development"},
		},
		Tags: []openswag.Tag{
			{Name: "Users", Description: "User management endpoints"},
		},
		UI: openswag.UIConfig{
			Theme:       "purple",
			DarkMode:    true,
			ShowSidebar: true,
		},
	})

	// Register endpoints
	docs.AddAll(CreateUserDoc, GetUserDoc, ListUsersDoc)

	// Setup routes
	mux := http.NewServeMux()
	docs.Mount(mux, "/docs")

	// API routes
	mux.HandleFunc("POST /users", createUser)
	mux.HandleFunc("GET /users", listUsers)
	mux.HandleFunc("GET /users/{id}", getUser)

	log.Println("Server running on http://localhost:8080")
	log.Println("Docs available at http://localhost:8080/docs/")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
