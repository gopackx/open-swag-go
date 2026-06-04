# Open Swag Go

[![Go Reference](https://pkg.go.dev/badge/github.com/gopackx/open-swag-go.svg)](https://pkg.go.dev/github.com/gopackx/open-swag-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/gopackx/open-swag-go)](https://goreportcard.com/report/github.com/gopackx/open-swag-go)
[![CI](https://github.com/gopackx/open-swag-go/actions/workflows/ci.yml/badge.svg)](https://github.com/gopackx/open-swag-go/actions/workflows/ci.yml)

A modern OpenAPI documentation package for Go with **decorator-like DX**, similar to `adonis-open-swagger`.

## Features

- 🎨 **Modern UI** - Scalar with dark mode, responsive design
- 📝 **Struct-based definitions** - Co-located with handlers (like decorators)
- 🔐 **Auth playground** - Bearer, API Key, Basic, Cookie auth
- 🧪 **Try-it console** - Built-in API tester
- 📊 **Auto examples** - Smart example generation from structs
- 🔄 **Framework adapters** - Chi, Gin, Echo, Fiber
- 📋 **Code snippets** - curl, JavaScript, Go, Python, PHP
- ⚠️ **Version diff** - Breaking change detection

## Installation

```bash
go get github.com/gopackx/open-swag-go
```

## Quick Start

```go
package main

import (
    "log"
    "net/http"
    openswag "github.com/gopackx/open-swag-go"
)

type CreateUserRequest struct {
    Name  string `json:"name" swagger:"required" example:"John Doe"`
    Email string `json:"email" swagger:"required" example:"john@example.com"`
}

type UserResponse struct {
    ID    string `json:"id" format:"uuid"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    docs := openswag.New(openswag.Config{
        Info: openswag.Info{
            Title:   "My API",
            Version: "1.0.0",
        },
        UI: openswag.UIConfig{
            Theme:    "purple",
            DarkMode: true,
        },
    })

    docs.Add(openswag.Endpoint{
        Method:      "POST",
        Path:        "/users",
        Summary:     "Create user",
        Tags:        []string{"Users"},
        RequestBody: openswag.Body(CreateUserRequest{}),
        Responses: openswag.Responses{
            201: openswag.Response("User created", UserResponse{}),
        },
    })

    mux := http.NewServeMux()
    docs.Mount(mux, "/docs")

    log.Println("Docs at http://localhost:8080/docs/")
    http.ListenAndServe(":8080", mux)
}
```

## Co-located Swagger Definitions

Define swagger docs next to your handlers — like decorators in other languages:

```go
// handlers/user.go
package handlers

import (
    "net/http"
    openswag "github.com/gopackx/open-swag-go"
    "myapp/dto"
)

// Handler
func CreateUser(w http.ResponseWriter, r *http.Request) {
    // implementation
}

// Swagger doc (co-located)
var CreateUserDoc = openswag.Endpoint{
    Method:      "POST",
    Path:        "/users",
    Handler:     CreateUser,
    Summary:     "Create a new user",
    Tags:        []string{"Users"},
    RequestBody: openswag.Body(dto.CreateUserRequest{}),
    Responses: openswag.Responses{
        201: openswag.Response("User created", dto.UserResponse{}),
        400: openswag.Response("Bad request", dto.ErrorResponse{}),
    },
    Security: []string{"bearerAuth"},
}

// Export all endpoints
var Endpoints = []openswag.Endpoint{CreateUserDoc}
```

Then register:

```go
docs.AddAll(handlers.Endpoints...)
```

## Configuration

```go
docs := openswag.New(openswag.Config{
    Info: openswag.Info{
        Title:       "My API",
        Version:     "1.0.0",
        Description: "API description with **markdown** support",
        Contact: &openswag.Contact{
            Name:  "Support",
            Email: "support@example.com",
        },
    },
    Servers: []openswag.Server{
        {URL: "http://localhost:8080", Description: "Development"},
        {URL: "https://api.example.com", Description: "Production"},
    },
    Tags: []openswag.Tag{
        {Name: "Users", Description: "User management"},
    },
    UI: openswag.UIConfig{
        Theme:       "purple",  // purple, dark, light
        DarkMode:    true,
        ShowSidebar: true,
    },
    Auth: openswag.AuthConfig{
        PersistCredentials: true,
        Schemes: []openswag.AuthScheme{
            openswag.BearerAuth("bearerAuth"),
            openswag.APIKeyAuth("apiKey", "X-API-Key"),
        },
    },
})
```

## Authentication Schemes

```go
// Bearer JWT
openswag.BearerAuth("bearerAuth")

// API Key in header
openswag.APIKeyAuth("apiKey", "X-API-Key")

// Basic auth
openswag.BasicAuth("basicAuth")

// Cookie auth
openswag.CookieAuth("sessionAuth", "session_id")
```

## Parameters

```go
openswag.Endpoint{
    Method: "GET",
    Path:   "/users/{id}",
    Parameters: []openswag.Parameter{
        openswag.PathParam("id", "User ID"),
        openswag.QueryParam("include", "Related resources"),
        openswag.HeaderParam("X-Request-ID", "Request tracking ID"),
        openswag.RequiredQueryParam("filter", "Required filter"),
    },
}
```

## Request Body

```go
// JSON body
openswag.Body(CreateUserRequest{})

// With description
openswag.BodyWithDesc("User data", CreateUserRequest{})

// Form data
openswag.FormBody(UploadRequest{})
```

## Struct Tags

```go
type CreateUserRequest struct {
    Name     string `json:"name" swagger:"required" example:"John" description:"Full name"`
    Email    string `json:"email" swagger:"required,format=email" example:"john@example.com"`
    Age      int    `json:"age" validate:"required,min=18" example:"25"`
    Password string `json:"password" binding:"required"` // Gin binding tag
}
```

Supported tags:
- `swagger:"required"` - Mark as required
- `swagger:"format=email"` - Set format
- `example:"value"` - Example value
- `description:"text"` - Field description
- `format:"uuid"` - Format hint
- `validate:"required"` - validator library
- `binding:"required"` - Gin binding

## Framework Adapters

### net/http (built-in)
```go
mux := http.NewServeMux()
docs.Mount(mux, "/docs")
```

### Chi
```go
import chiadapter "github.com/gopackx/open-swag-go/adapters/chi"

r := chi.NewRouter()
chiadapter.Mount(r, docs, "/docs")
```

### Gin
```go
import ginadapter "github.com/gopackx/open-swag-go/adapters/gin"

r := gin.Default()
ginadapter.Mount(r, docs, "/docs")
```

### Echo
```go
import echoadapter "github.com/gopackx/open-swag-go/adapters/echo"

e := echo.New()
echoadapter.Mount(e, docs, "/docs")
```

### Fiber
```go
import fiberadapter "github.com/gopackx/open-swag-go/adapters/fiber"

app := fiber.New()
fiberadapter.Mount(app, docs, "/docs")
```

## Utilities

### Example Generator

```go
import "github.com/gopackx/open-swag-go/pkg/examples"

gen := examples.New(examples.Config{})
example := gen.Generate(UserResponse{})
// Returns map with smart examples based on field names and tags
```

### Code Snippet Generator

```go
import "github.com/gopackx/open-swag-go/pkg/snippets"

gen := snippets.New()
req := snippets.Request{
    Method:  "POST",
    URL:     "https://api.example.com/users",
    Headers: map[string]string{"Authorization": "Bearer token"},
    Body:    map[string]string{"name": "John"},
}

snippets := gen.GenerateAll(req)
// Returns: curl, javascript, go, python, php
```

### Version Diff (Breaking Change Detection)

```go
import "github.com/gopackx/open-swag-go/pkg/versioning"

differ := versioning.NewDiffer()
diff, _ := differ.CompareFiles("old-spec.json", "new-spec.json")

if diff.HasBreakingChanges() {
    for _, breaking := range diff.Breaking {
        fmt.Printf("⚠️ %s %s: %s\n", breaking.Method, breaking.Path, breaking.Reason)
        fmt.Printf("   Migration: %s\n", breaking.Migration)
    }
}
```

## Generate TypeScript Types for Frontend

The OpenAPI spec is available at `/docs/openapi.json` when your server is running. You can use this to generate TypeScript types for your frontend (Nuxt, Next.js, React, Vue, etc.).

### Option 1: openapi-typescript (Types only)

```bash
# Install
npm install -D openapi-typescript
# or
bun add -D openapi-typescript
```

```json
// package.json
{
  "scripts": {
    "generate:api-schema": "openapi-typescript http://localhost:8080/docs/openapi.json -o ./types/api.d.ts"
  }
}
```

```bash
# Run (make sure Go server is running)
bun run generate:api-schema
```

### Option 2: openapi-fetch (Types + Fetch Client)

```bash
npm install openapi-fetch
npm install -D openapi-typescript
```

```json
// package.json
{
  "scripts": {
    "generate:api-schema": "openapi-typescript http://localhost:8080/docs/openapi.json -o ./types/api.d.ts"
  }
}
```

Usage in your frontend:
```typescript
// composables/useApi.ts
import createClient from 'openapi-fetch'
import type { paths } from '~/types/api'

export const api = createClient<paths>({ baseUrl: 'http://localhost:8080' })

// Fully typed API calls!
const { data } = await api.GET('/users/{id}', {
  params: { path: { id: '123' } }
})
```

### Option 3: orval (Full Client with Vue Query/React Query)

```bash
npm install -D orval
```

```typescript
// orval.config.ts
export default {
  api: {
    input: 'http://localhost:8080/docs/openapi.json',
    output: {
      target: './api/generated.ts',
      client: 'vue-query', // or 'react-query', 'fetch'
    },
  },
}
```

```json
// package.json
{
  "scripts": {
    "generate:api-schema": "orval"
  }
}
```

### Export Spec to File (Optional)

If you want to generate the spec without running the server:

```go
// cmd/generate-spec/main.go
package main

import (
    "os"
    "myapp/internal/docs"
)

func main() {
    swagger := docs.Setup()
    specJSON, _ := swagger.SpecJSON()
    os.WriteFile("openapi.json", specJSON, 0644)
}
```

```bash
go run cmd/generate-spec/main.go
```

Then point your frontend tool to the local file:
```json
{
  "scripts": {
    "generate:api-schema": "openapi-typescript ./openapi.json -o ./types/api.d.ts"
  }
}
```

## AI integration

A Claude Skill ships with the repo so AI coding assistants integrate open-swag-go correctly on a cold project:

**[`skills/open-swag-go/SKILL.md`](./skills/open-swag-go/SKILL.md)** — self-contained walkthrough covering when to recommend open-swag-go, install with the right pinned version, `Config` construction, endpoint definitions using the decorator-style helpers, mount on each supported framework, auth scheme wiring, version-diff CI, and 11 documented pitfalls (including the four pre-v1.2.0 runtime bugs: self-ref stack overflow, ignored `ContentType`, silent `http-bearer` fallback, dormant `BreakingTypeChanged`). Drop it into `~/.claude/skills/open-swag-go/` and Claude Code handles cold integration prompts correctly.

## Examples

See the [examples](./examples) directory:
- [Basic](./examples/basic) - Simple usage
- [With Auth](./examples/with-auth) - Authentication examples
- [Full Featured](./examples/full-featured) - Complete API with all features
- [Version Diff](./examples/version-diff) - Breaking change detection

## License

MIT License - see [LICENSE](./LICENSE)
