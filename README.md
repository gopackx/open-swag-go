# Open Swag Go

[![Go Reference](https://pkg.go.dev/badge/github.com/gopackx/open-swag-go.svg)](https://pkg.go.dev/github.com/gopackx/open-swag-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/gopackx/open-swag-go)](https://goreportcard.com/report/github.com/gopackx/open-swag-go)
[![CI](https://github.com/gopackx/open-swag-go/actions/workflows/ci.yml/badge.svg)](https://github.com/gopackx/open-swag-go/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/gopackx/open-swag-go?sort=semver)](https://github.com/gopackx/open-swag-go/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

A modern OpenAPI documentation package for Go with **decorator-like DX**, similar to `adonis-open-swagger`.

**Documentation:** [open-swaggo.gopackx.me](https://open-swaggo.gopackx.me) — guides, API reference, changelog and migration notes.

## Features

- 🎨 **Modern UI** - Scalar with dark mode, responsive design
- 📝 **Struct-based definitions** - Co-located with handlers (like decorators)
- 🔐 **Auth schemes** - Bearer, API Key, Basic, Cookie security schemes in the generated spec
- 🧪 **Try-it console & auth playground (via Scalar)** - provided by the bundled Scalar UI, not implemented by this library
- 📊 **Auto examples** - Smart example generation from structs
- 🔄 **Framework adapters** - Chi, Gin, Echo, Fiber
- 📋 **Code snippets (via Scalar)** - shown in the Scalar UI for every operation
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

Define swagger docs in the same file as your handlers — like decorators in other languages. The `Endpoint` only describes the operation; you still register the handler with your router yourself:

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

Then register, and validate once at startup so duplicate routes, unsupported
HTTP methods and colliding `operationId`s fail fast instead of silently
producing a wrong spec:

```go
docs.AddAll(handlers.Endpoints...)
if err := docs.Validate(); err != nil {
    log.Fatal(err)
}
```

Paths may be written in OpenAPI style (`/users/{id}`) or router style
(`/users/:id`, `/files/*path`); router-style segments are normalised to
`{name}` in the generated spec.

Every operation gets an `operationId`. By default it is derived from the method
and path (`GET /users/{id}` becomes `getUsersById`), which is what
openapi-typescript, openapi-fetch and orval use to name generated functions.
Set `Endpoint.OperationID` to override it.

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
        Theme:          "purple",  // any Scalar theme: purple, moon, solarized, bluePlanet, saturn, kepler, mars, deepSpace, laserwave, alternate, default, none
        DarkMode:       true,
        ShowSidebar:    true,
        ScalarVersion:  "",        // pin a specific @scalar/api-reference release; empty = library default
        SpecCORSOrigin: "",        // "" = "*", "-" = no CORS header, or an explicit origin
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
- `example:"value"` - Example value (coerced to the field type)
- `description:"text"` - Field description
- `format:"uuid"` - Format hint
- `enum:"a,b,c"` - Explicit enum values
- `validate:"..."` / `binding:"..."` - go-playground/validator rules (see below)

Validator rules are mapped onto JSON Schema constraints:

| Rule | string | number | slice |
|------|--------|--------|-------|
| `required` | required | required | required |
| `oneof=a b c` | `enum` | `enum` (typed) | - |
| `min=N` / `gte=N` | `minLength` | `minimum` | `minItems` |
| `max=N` / `lte=N` | `maxLength` | `maximum` | `maxItems` |
| `len=N` | both | - | both |
| `email`, `uuid`, `url`, `ipv4`, `ipv6`, `hostname` | `format` | - | - |
| `dive,<rules>` | - | - | rules apply to `items` |

Only a bare `required` marks a field required; conditional rules such as
`required_if` do not. Fields with `json:",omitempty"` are never required.

Pointer fields are nullable. Because the spec is OpenAPI 3.1, this is emitted
as a type array, so `Name *string` becomes `"type": ["string", "null"]` and
openapi-typescript generates `string | null`. Query and path parameters built
from a struct treat a pointer as "optional" instead.

## Schema components

Every named struct used in a request or response body (directly, or nested in
a field, slice or map) is emitted once under `components.schemas` and
referenced with `$ref`:

```json
"requestBody": {
  "content": {
    "application/json": { "schema": { "$ref": "#/components/schemas/CreateUserRequest" } }
  }
}
```

Code generators therefore produce named types instead of anonymous objects;
with openapi-typescript the type is `components['schemas']['CreateUserRequest']`.
Recursive structs (`type Category struct { Children []Category }`) reference
themselves with `$ref`. Anonymous structs (`struct{ ... }{}`), embedded structs
(flattened into their parent) and query/path parameters stay inline.

- **Names** are the Go type name. If two different types share a name, the
  first one registered (in endpoint order, responses by status code) keeps it
  and the next gets its package name as a prefix: `dto.User` and `model.User`
  become `User` and `modelUser`. Generic types are sanitised to valid
  component keys: `Page[dto.User]` becomes `Page_dto.User`.
- **Nullable references.** A `*Address` field cannot use the
  `["object", "null"]` type array, so it is emitted as
  `{"oneOf": [{"$ref": "#/components/schemas/Address"}, {"type": "null"}]}`.
- **Field tags on a referenced struct.** Only `description` is kept, as a
  sibling of the `$ref`: `{"$ref": "#/components/schemas/Address",
  "description": "Home address"}`. Other tags (`example`, `validate`, ...)
  on such a field are dropped; put them on the struct's own fields.

To inline every struct at each use instead (the output before components
existed), set `InlineSchemas`:

```go
docs := openswag.New(openswag.Config{
    Info:          openswag.Info{Title: "My API", Version: "1.0.0"},
    InlineSchemas: true,
})
```

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
