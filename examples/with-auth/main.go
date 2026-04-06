package main

import (
	"encoding/json"
	"log"
	"net/http"

	openswag "github.com/gopackx/open-swag-go"
	"github.com/gopackx/open-swag-go/pkg/auth"
	"github.com/gopackx/open-swag-go/pkg/spec"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{Token: "jwt-token", ExpiresIn: 3600})
}

func register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UserResponse{ID: "1", Email: "user@example.com"})
}

func getProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserResponse{ID: "1", Email: "user@example.com", Name: "John"})
}

var LoginDoc = openswag.Endpoint{
	Method:  "POST",
	Path:    "/auth/login",
	Summary: "User login",
	Tags:    []string{"Auth"},
	RequestBody: &openswag.RequestBody{
		Description: "Login credentials",
		Required:    true,
		Schema:      LoginRequest{},
	},
	Responses: map[int]openswag.Response{
		200: {Description: "Login successful", Schema: TokenResponse{}},
		401: {Description: "Invalid credentials", Schema: ErrorResponse{}},
	},
}

var RegisterDoc = openswag.Endpoint{
	Method:  "POST",
	Path:    "/auth/register",
	Summary: "User registration",
	Tags:    []string{"Auth"},
	RequestBody: &openswag.RequestBody{
		Description: "Registration data",
		Required:    true,
		Schema:      LoginRequest{},
	},
	Responses: map[int]openswag.Response{
		201: {Description: "User created", Schema: UserResponse{}},
		400: {Description: "Invalid data", Schema: ErrorResponse{}},
	},
}

var ProfileDoc = openswag.Endpoint{
	Method:   "GET",
	Path:     "/users/me",
	Summary:  "Get current user profile",
	Tags:     []string{"Users"},
	Security: []string{"bearerAuth"},
	Responses: map[int]openswag.Response{
		200: {Description: "Profile retrieved", Schema: UserResponse{}},
	},
}

func main() {
	docs := openswag.New(openswag.Config{
		Info: openswag.Info{
			Title:       "Auth API",
			Version:     "1.0.0",
			Description: "API with authentication examples",
		},
		Servers: []openswag.Server{
			{URL: "http://localhost:8080", Description: "Development"},
		},
		Tags: []openswag.Tag{
			{Name: "Auth", Description: "Authentication endpoints"},
			{Name: "Users", Description: "User endpoints"},
		},
		UI: openswag.UIConfig{
			Theme:       "purple",
			DarkMode:    true,
			ShowSidebar: true,
		},
	})

	bearerScheme := auth.BearerAuth("JWT authentication")
	apiKeyScheme := auth.APIKeyHeader("X-API-Key", "API key authentication")

	openapi := docs.BuildSpec()
	openapi.AddSecurityScheme("bearerAuth", &spec.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  bearerScheme.Description,
	})
	openapi.AddSecurityScheme("apiKey", &spec.SecurityScheme{
		Type:        "apiKey",
		Name:        apiKeyScheme.Name,
		In:          string(apiKeyScheme.In),
		Description: apiKeyScheme.Description,
	})

	docs.AddAll(LoginDoc, RegisterDoc, ProfileDoc)

	mux := http.NewServeMux()
	docs.Mount(mux, "/docs")

	mux.HandleFunc("POST /auth/login", login)
	mux.HandleFunc("POST /auth/register", register)
	mux.HandleFunc("GET /users/me", getProfile)

	log.Println("Server running on http://localhost:8080")
	log.Println("Docs at http://localhost:8080/docs/")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
