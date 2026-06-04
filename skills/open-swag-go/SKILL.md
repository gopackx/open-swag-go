# open-swag-go integration skill

---
name: open-swag-go
description: Use this skill when integrating gopackx/open-swag-go (a Go library that generates an OpenAPI 3.x spec from struct-tagged endpoint definitions and serves a Scalar-powered docs UI) into a Go project. Triggers include any of these in a Go context — adding "OpenAPI docs" / "Swagger UI" / "API documentation" / "interactive API explorer" to a Go service; defining endpoints with decorator-style co-located doc structs; generating schemas from existing DTO/request/response structs (json/form/swagger/validate/binding tags); mounting docs on net/http, Chi, Gin, Echo, or Fiber; configuring auth schemes (Bearer JWT, API key, basic, cookie, OAuth2) for the spec and try-it console; exporting openapi.json to feed openapi-typescript / openapi-fetch / orval frontends; detecting breaking changes between two spec versions. Use also for upgrading an existing open-swag-go deployment between versions (v1.1.x → v1.2.0 has one renamed type, see "Common pitfalls"). Do NOT use for: hand-writing OpenAPI YAML, swag/swaggo (a different annotation-comment library), non-Go projects, or runtime request validation (this library generates docs, not middleware).
---

`github.com/gopackx/open-swag-go` turns Go struct definitions into an OpenAPI 3.x spec and serves a Scalar-powered interactive docs UI. The core depends only on stdlib; framework adapters are opt-in subpackages.

This skill walks you through integrating it correctly the first time and lists the pitfalls that cause most early issues.

---

## Step 0 — Does the user actually want open-swag-go?

Confirm fit before suggesting it:

✅ **Good fit**
- Writing a Go service (any framework) that needs interactive API docs without hand-writing OpenAPI YAML.
- Has existing request/response DTO structs and wants the schema generated from them automatically.
- Wants the doc definitions co-located with handlers (decorator-style) rather than annotation-comment scanning.
- Already on net/http, Chi, Gin, Echo, or Fiber.
- Plans to feed `openapi.json` to a TypeScript codegen tool (openapi-typescript, openapi-fetch, orval) on the frontend.
- Wants breaking-change detection between two versioned specs in CI.

❌ **Not the right tool**
- They want comment/annotation scanning (`// @Summary …`) → recommend `swaggo/swag`.
- They want runtime request/response validation → recommend `kin-openapi` or `go-playground/validator` directly.
- They want to hand-author OpenAPI YAML → walk away.
- Non-Go project → walk away.

---

## Step 1 — Install

Pin to `v1.2.0` or later — earlier releases have four bugs that produce wrong specs at runtime (self-referential struct stack overflow, ignored `RequestBody.ContentType`, silent `http-bearer` fallback for unknown security schemes, dormant `BreakingTypeChanged` in version diff). See "Common pitfalls" for details.

```bash
go get github.com/gopackx/open-swag-go@v1.2.0
```

No separate adapter install is needed — the framework adapters live in subpackages of the same module (`/adapters/chi`, `/adapters/gin`, `/adapters/echo`, `/adapters/fiber`, `/adapters/nethttp`) and come along automatically.

---

## Step 2 — Build the `Config`

The minimum is `Info.Title` + `Info.Version`. Everything else is optional but usually worth setting:

```go
import openswag "github.com/gopackx/open-swag-go"

cfg := openswag.Config{
    Info: openswag.Info{
        Title:       "Pet Store API",
        Version:     "1.0.0",
        Description: "Markdown **supported** here.",
        Contact:     &openswag.Contact{Name: "API Team", Email: "api@example.com"},
        License:     &openswag.License{Name: "MIT", URL: "https://opensource.org/licenses/MIT"},
    },
    Servers: []openswag.Server{
        {URL: "http://localhost:8080", Description: "Local"},
        {URL: "https://api.example.com", Description: "Production"},
    },
    Tags: []openswag.Tag{
        {Name: "Pets", Description: "Pet management"},
    },
    UI: openswag.UIConfig{
        Theme:       "purple", // purple / dark / light
        DarkMode:    true,
        ShowSidebar: true,
        Layout:      "modern",
    },
    Auth: openswag.AuthConfig{
        PersistCredentials: true,
        Schemes: []openswag.AuthScheme{
            openswag.BearerAuth(openswag.SecurityBearerAuth),
            openswag.APIKeyAuth("apiKey", "X-API-Key"),
        },
    },
}

docs := openswag.New(cfg)
```

Common mistakes here:

- **Forgetting `Auth.Schemes` when endpoints reference a custom scheme name.** v1.2.0 stopped silently coercing unknown scheme names to `http-bearer`. If you write `Security: []string{"myCustomScheme"}` you must also register `myCustomScheme` in `Config.Auth.Schemes`, otherwise it's omitted from `components.securitySchemes`. Predefined `SecurityBearerAuth` / `SecurityBasicAuth` / `SecurityApiKey` / `SecurityApiKeyQuery` / `SecurityCookieAuth` / `SecurityOAuth2` constants still auto-register their built-in schemes.
- **Setting `UI.Theme` to an unknown string** — silently falls back to `purple`. Stick to the documented values.
- **Putting docs auth on `Auth`** — that's for the API's security schemes. To password-protect the docs UI itself, use `DocsAuth` (basic auth or `?key=` query param).

---

## Step 3 — Define endpoints

The decorator-style helpers (added in v1.2.0) make this concise:

```go
type CreateUserRequest struct {
    Name  string `json:"name" swagger:"required" example:"John Doe"`
    Email string `json:"email" swagger:"required,format=email" example:"john@example.com"`
    Age   int    `json:"age" validate:"min=18" example:"25"`
}

type UserResponse struct {
    ID    string `json:"id" format:"uuid"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type ErrorResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

var CreateUserDoc = openswag.Endpoint{
    Method:      "POST",
    Path:        "/users",
    Summary:     "Create a new user",
    Tags:        []string{"Users"},
    RequestBody: openswag.BodyWithDesc("User payload", CreateUserRequest{}),
    Responses: openswag.Responses{
        201: openswag.Response("User created", UserResponse{}),
        400: openswag.Response("Invalid input", ErrorResponse{}),
    },
    Security: []string{openswag.SecurityBearerAuth},
}

var GetUserDoc = openswag.Endpoint{
    Method:  "GET",
    Path:    "/users/{id}",
    Summary: "Get user by ID",
    Tags:    []string{"Users"},
    Parameters: []openswag.Parameter{
        openswag.PathParam("id", "User ID"),
    },
    Responses: openswag.Responses{
        200: openswag.Response("User found", UserResponse{}),
        404: openswag.Response("Not found", ErrorResponse{}),
    },
    Security: []string{openswag.SecurityBearerAuth},
}

docs.AddAll(CreateUserDoc, GetUserDoc)
```

Quick reference for the helpers:

| Helper | What it builds |
|--------|----------------|
| `Body(T{})` | `*RequestBody`, required, `application/json` |
| `BodyWithDesc("desc", T{})` | Same, with description |
| `FormBody(T{})` | `multipart/form-data` body (file upload) |
| `FormURLEncodedBody(T{})` | `application/x-www-form-urlencoded` |
| `Response("desc", T{})` | `ResponseSpec` for the `Responses{}` map |
| `Response("desc")` | Empty response (e.g. 204 No Content) |
| `PathParam(name, desc)` | Required string path param |
| `QueryParam(name, desc)` | Optional string query param |
| `RequiredQueryParam(name, desc)` | Required string query param |
| `HeaderParam(name, desc)` | Optional string header param |
| `CookieParam(name, desc)` | Optional string cookie param |
| `BearerAuth(name)` | JWT bearer `AuthScheme` for `Config.Auth.Schemes` |
| `APIKeyAuth(name, header)` | apiKey-in-header `AuthScheme` |
| `APIKeyQueryAuth(name, param)` | apiKey-in-query `AuthScheme` |
| `BasicAuth(name)` | HTTP basic `AuthScheme` |
| `CookieAuth(name, cookieName)` | apiKey-in-cookie `AuthScheme` |

### Typed query / path parameters from a struct

When you have several query or path params, define them as a struct and let the library reflect over the field tags instead of writing `Parameters` manually:

```go
type ListUsersQuery struct {
    Page    int    `form:"page" description:"Page number" example:"1"`
    PerPage int    `form:"per_page" description:"Items per page" example:"20"`
    Search  string `form:"search" description:"Search term"`
}

type UserPathParams struct {
    ID string `param:"id" description:"User ID" example:"usr_123"`
}

openswag.Endpoint{
    Method:      "GET",
    Path:        "/users/{id}",
    PathParams:  UserPathParams{},
    QueryParams: ListUsersQuery{},
    // ...
}
```

Tags the library understands on DTO fields: `json`, `form`, `query`, `path`, `param`, `swagger:"required,format=…"`, `validate:"required"`, `binding:"required"`, `example:"…"` (type-coerced), `description:"…"`, `format:"uuid|email|date-time|…"`.

---

## Step 4 — Mount the docs

### net/http (built-in)

```go
mux := http.NewServeMux()
docs.Mount(mux, "/docs")
// UI:  http://localhost:8080/docs/
// JSON: http://localhost:8080/docs/openapi.json
http.ListenAndServe(":8080", mux)
```

### Chi / Gin / Echo / Fiber

```go
import chiadapter "github.com/gopackx/open-swag-go/adapters/chi"
r := chi.NewRouter()
chiadapter.Mount(r, docs, "/docs")

import ginadapter "github.com/gopackx/open-swag-go/adapters/gin"
ginadapter.Mount(ginEngine, docs, "/docs")

import echoadapter "github.com/gopackx/open-swag-go/adapters/echo"
echoadapter.Mount(echoEngine, docs, "/docs")

import fiberadapter "github.com/gopackx/open-swag-go/adapters/fiber"
fiberadapter.Mount(fiberApp, docs, "/docs")
```

### Protect the docs UI (optional)

```go
DocsAuth: &openswag.DocsAuth{
    Enabled:  true,
    Username: "admin",
    Password: os.Getenv("DOCS_PASS"),
    // OR use an API key query param instead:
    // APIKey: os.Getenv("DOCS_KEY"), // access via ?key=...
},
```

---

## Step 5 — Export `openapi.json` for the frontend

The spec is served at `/docs/openapi.json` while the server runs. To generate without running the server:

```go
swagger := openswag.New(cfg)
swagger.AddAll(endpoints...)
specJSON, _ := swagger.SpecJSON()
_ = os.WriteFile("openapi.json", specJSON, 0644)
```

Then point your frontend tool at it:

```json
{
  "scripts": {
    "generate:api": "openapi-typescript ./openapi.json -o ./types/api.d.ts"
  }
}
```

Works the same with `openapi-fetch`, `orval`, or any other OpenAPI-3 consumer.

---

## Step 6 — Version diff (optional, useful in CI)

Compare two saved specs and fail the build on breaking changes:

```go
import "github.com/gopackx/open-swag-go/pkg/versioning"

diff, _ := versioning.NewDiffer().CompareFiles("openapi.previous.json", "openapi.json")

if diff.HasBreakingChanges() {
    for _, b := range diff.Breaking {
        fmt.Printf("BREAKING: %s %s — %s (migrate: %s)\n",
            b.Method, b.Path, b.Reason, b.Migration)
    }
    os.Exit(1)
}
```

What the differ detects as breaking: endpoint removed, request body removed, required body added, new required field, **field type changed**, **parameter type changed**, parameter removed, new required parameter, response code removed. (The two bolded items were dormant before v1.2.0 — the constant `BreakingTypeChanged` existed but was never produced.)

---

## Common pitfalls

These are the issues real users hit on the first integration:

1. **Pre-v1.2.0 self-referential struct stack overflow.** Types like
   `type Category struct { Children []Category }` recurse forever in
   `schema.fromReflectType` on v1.1.x and earlier. Always pin
   `v1.2.0` or later.

2. **Pre-v1.2.0 `RequestBody.ContentType` was ignored.** File-upload
   endpoints that set `ContentType: "multipart/form-data"` were
   silently advertised as `application/json`, breaking downstream
   client generators. Upgrade. Prefer the `FormBody()` /
   `FormURLEncodedBody()` helpers over hand-built `*RequestBody`
   literals.

3. **Pre-v1.2.0 silent `http-bearer` fallback.** Any unknown name in
   `Endpoint.Security` (typos, custom scheme names) was silently
   coerced into `http-bearer` JWT. After v1.2.0, unknown names are
   *omitted* — register them via `Config.Auth.Schemes`. Migration:

   ```go
   Auth: openswag.AuthConfig{
       Schemes: []openswag.AuthScheme{
           openswag.BearerAuth(openswag.SecurityBearerAuth),
           openswag.APIKeyAuth("apiKey", "X-API-Key"),
           openswag.CookieAuth("sessionAuth", "session_id"),
       },
   },
   ```

4. **v1.2.0 renamed `Response` struct → `ResponseSpec`.** `Response`
   is now a constructor function (`Response("desc", T{}) ResponseSpec`).
   Field-keyed literals inside an `openswag.Responses{ ... }` map keep
   compiling via composite-literal type inference. Bare
   `openswag.Response{...}` literals need to become either
   `openswag.Response("desc", T{})` or `openswag.ResponseSpec{...}`.

5. **`example:"42"` showing as the string `"42"` in the spec.** Pre-
   v1.1.0 bug — example values weren't type-coerced. Upgrade.

6. **`json:",omitempty"` field marked as required.** Pre-v1.1.0 the
   `omitempty` flag was ignored when deciding required-ness. Upgrade.

7. **`$ref` pointers vanishing.** Pre-v1.1.0 `convertSchema` dropped
   the `Ref` field. Upgrade.

8. **Embedded struct fields missing.** Anonymous (embedded) struct
   fields are flattened into the parent schema since v1.1.0. On older
   versions they were silently ignored.

9. **Custom field name when only `json:",omitempty"` is set.** The
   empty-name part now correctly falls back to the Go field name
   (since v1.1.0).

10. **Docs UI shows nothing / 404.** `Mount(mux, "/docs")` registers
    on `"/docs/"` (trailing slash). Browsing to `"/docs"` without the
    slash on net/http's default ServeMux issues a 301 to `"/docs/"` —
    fine for browsers, sometimes confusing in curl tests. Always test
    with the trailing slash.

11. **`docs.BuildSpec()` cached.** Calling `Add`/`AddAll` invalidates
    the cache, but if you mutate `openapi.Components` directly after
    `BuildSpec()` and then add another endpoint, your mutations are
    lost on the rebuild. Prefer `Config.Auth.Schemes` over post-build
    `openapi.AddSecurityScheme(...)`.

---

## Quick-start checklist (paste-ready)

When the user says "add OpenAPI docs to my Go service," confirm these decisions before writing code:

1. Framework? (net/http / chi / gin / echo / fiber)
2. Mount path? (default `/docs`)
3. Do they have existing DTO structs to reflect over, or starting from scratch?
4. Auth on the API? (bearer / api key / basic / cookie / oauth2 / none)
5. Auth on the docs UI itself? (none / basic / API key query)
6. Will they regenerate frontend types from `openapi.json`? (informs whether to add a `cmd/generate-spec` binary)
7. Do they want version-diff in CI? (informs whether to save `openapi.json` as a checked-in artifact)

Once you have these, generate the `openswag.Config{...}` literal, the endpoint `openswag.Endpoint{...}` definitions using the decorator helpers, and the one-line `docs.Mount(...)` call. That's the whole integration.

---

## Reference

- Repo: <https://github.com/gopackx/open-swag-go>
- Docs site: <https://open-swag-go.dev> (or whichever URL the user's deploy renders).
- Examples in the repo: `examples/basic`, `examples/with-auth`, `examples/full-featured`, `examples/version-diff`.
- Changelog: check the docs site `Changelog` page first when something behaves unexpectedly — every release since v1.1.0 documents the fix-by-fix delta.
