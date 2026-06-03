package main

import (
	"encoding/json"
	"log"
	"net/http"

	openswag "github.com/gopackx/open-swag-go"
)

// DTOs
type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	CategoryID  string  `json:"category_id"`
}

type ProductResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	CategoryID  string  `json:"category_id"`
	CreatedAt   string  `json:"created_at"`
}

type PaginatedProducts struct {
	Data       []ProductResponse `json:"data"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	Total      int               `json:"total"`
	TotalPages int               `json:"total_pages"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Query/Path parameter structs (NEW FEATURE)
type ListProductsQuery struct {
	Page     int    `form:"page" description:"Page number" example:"1"`
	PerPage  int    `form:"per_page" description:"Items per page" example:"20"`
	Category string `form:"category" description:"Filter by category ID"`
	Search   string `form:"search" description:"Search term"`
}

type ProductPathParams struct {
	ID string `param:"id" description:"Product ID" example:"prod_123"`
}

// Handlers
func createProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ProductResponse{ID: "1", Name: "Product"})
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProductResponse{ID: "1", Name: "Product"})
}

func listProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PaginatedProducts{Data: []ProductResponse{}, Page: 1, PerPage: 20})
}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProductResponse{ID: "1", Name: "Updated"})
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Endpoint definitions - use predefined security constants
var CreateProductDoc = openswag.Endpoint{
	Method:      "POST",
	Path:        "/products",
	Summary:     "Create a new product",
	Description: "Create a new product in the catalog",
	Tags:        []string{"Products"},
	RequestBody: openswag.BodyWithDesc("Product data", CreateProductRequest{}),
	Responses: openswag.Responses{
		201: openswag.Response("Product created", ProductResponse{}),
		400: openswag.Response("Invalid request", ErrorResponse{}),
		401: openswag.Response("Unauthorized", ErrorResponse{}),
		500: openswag.Response("Server error", ErrorResponse{}),
	},
	Security: []string{openswag.SecurityBearerAuth}, // Use predefined constant
}

var GetProductDoc = openswag.Endpoint{
	Method:      "GET",
	Path:        "/products/{id}",
	Summary:     "Get product by ID",
	Description: "Retrieve a single product by its ID",
	Tags:        []string{"Products"},
	PathParams:  ProductPathParams{}, // Using struct instead of manual Parameters
	Responses: openswag.Responses{
		200: openswag.Response("Product found", ProductResponse{}),
		404: openswag.Response("Product not found", ErrorResponse{}),
	},
}

var ListProductsDoc = openswag.Endpoint{
	Method:      "GET",
	Path:        "/products",
	Summary:     "List products",
	Description: "Get a paginated list of products with optional filtering",
	Tags:        []string{"Products"},
	QueryParams: ListProductsQuery{}, // Using struct instead of manual Parameters
	Responses: openswag.Responses{
		200: openswag.Response("Products list", PaginatedProducts{}),
	},
}

var UpdateProductDoc = openswag.Endpoint{
	Method:      "PUT",
	Path:        "/products/{id}",
	Summary:     "Update product",
	Description: "Update an existing product",
	Tags:        []string{"Products"},
	PathParams:  ProductPathParams{},
	RequestBody: openswag.BodyWithDesc("Updated product data", CreateProductRequest{}),
	Responses: openswag.Responses{
		200: openswag.Response("Product updated", ProductResponse{}),
		400: openswag.Response("Invalid request", ErrorResponse{}),
		404: openswag.Response("Product not found", ErrorResponse{}),
	},
	Security: []string{openswag.SecurityBearerAuth},
}

var DeleteProductDoc = openswag.Endpoint{
	Method:      "DELETE",
	Path:        "/products/{id}",
	Summary:     "Delete product",
	Description: "Delete a product from the catalog",
	Tags:        []string{"Products"},
	PathParams:  ProductPathParams{},
	Responses: openswag.Responses{
		204: openswag.Response("Product deleted"),
		404: openswag.Response("Product not found", ErrorResponse{}),
	},
	Security:   []string{openswag.SecurityBearerAuth},
	Deprecated: false,
}

func main() {
	docs := openswag.New(openswag.Config{
		Info: openswag.Info{
			Title:       "E-Commerce API",
			Version:     "2.0.0",
			Description: "Full-featured e-commerce API with products, categories, and orders",
			Contact: &openswag.Contact{
				Name:  "API Team",
				Email: "api@example.com",
				URL:   "https://example.com/support",
			},
			License: &openswag.License{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
		},
		Servers: []openswag.Server{
			{URL: "http://localhost:8080", Description: "Development"},
			{URL: "https://staging.example.com", Description: "Staging"},
			{URL: "https://api.example.com", Description: "Production"},
		},
		Tags: []openswag.Tag{
			{Name: "Products", Description: "Product management"},
			{Name: "Categories", Description: "Category management"},
			{Name: "Orders", Description: "Order management"},
		},
		UI: openswag.UIConfig{
			Theme:       "purple",
			DarkMode:    true,
			ShowSidebar: true,
			Layout:      "modern",
		},
		// Protect docs UI (optional)
		DocsAuth: &openswag.DocsAuth{
			Enabled: true,
			APIKey:  "my-secret-key", // Access via ?key=my-secret-key
		},
		// Security schemes are auto-generated based on endpoint usage!
		// Just use: Security: []string{openswag.SecurityBearerAuth}
	})

	docs.AddAll(
		CreateProductDoc,
		GetProductDoc,
		ListProductsDoc,
		UpdateProductDoc,
		DeleteProductDoc,
	)

	mux := http.NewServeMux()
	docs.Mount(mux, "/docs")

	mux.HandleFunc("POST /products", createProduct)
	mux.HandleFunc("GET /products", listProducts)
	mux.HandleFunc("GET /products/{id}", getProduct)
	mux.HandleFunc("PUT /products/{id}", updateProduct)
	mux.HandleFunc("DELETE /products/{id}", deleteProduct)

	log.Println("Server running on http://localhost:8080")
	log.Println("Docs at http://localhost:8080/docs/")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
