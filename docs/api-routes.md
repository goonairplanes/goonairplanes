# 🔌 API Routes

This guide covers how to create and work with API routes in your GOA application.

## Creating Routes

### Basic Route
```go
// app/api/hello/route.go
package hello

import (
    "goonairplanes/core"
)

func init() {
    core.RegisterAPIHandler("/api/hello", "GET", Handler)
}

func Handler(ctx *core.APIContext) {
    ctx.Success(map[string]string{
        "message": "Hello!",
    }, 200)
}
```

### Route with Parameters
```go
// app/api/user/route.go
package user

import (
    "goonairplanes/core"
)

func init() {
    core.RegisterAPIHandler("/api/user/[id]", "GET", Handler)
}

func Handler(ctx *core.APIContext) {
    userID := ctx.Params["id"]
    ctx.Success(map[string]string{
        "id": userID,
        "name": "John Doe",
    }, 200)
}
```

### POST Route
```go
// app/api/create/route.go
package create

import (
    "goonairplanes/core"
)

func init() {
    core.RegisterAPIHandler("/api/create", "POST", Handler)
}

func Handler(ctx *core.APIContext) {
    var data map[string]interface{}
    err := ctx.ParseBody(&data)
    if err != nil {
        ctx.Error("Invalid request body", 400)
        return
    }
    
    ctx.Success(map[string]interface{}{
        "status": "success",
        "data": data,
    }, 201)
}
```

## Import System

To enable your API routes, import them in the main.go file:

```go
package main

import (
    "flag"
    "goonairplanes/core"
    "log"

    // Import your API routes packages
    _ "goonairplanes/app/api/hello"
    _ "goonairplanes/app/api/user"
    _ "goonairplanes/app/api/create"
)

func main() {
    // Application startup code
}
```

## Request Handling

### Reading Query Parameters
```go
func Handler(ctx *core.APIContext) {
    query := ctx.Request.URL.Query()
    page := query.Get("page")
    limit := query.Get("limit")
    // Use parameters
}
```

### Reading POST Data
```go
func Handler(ctx *core.APIContext) {
    var input struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }
    err := ctx.ParseBody(&input)
    if err != nil {
        ctx.Error("Invalid request body", 400)
        return
    }
    // Use input data
}
```

### File Upload
```go
func Handler(ctx *core.APIContext) {
    file, header, err := ctx.Request.FormFile("file")
    if err != nil {
        ctx.Error(err.Error(), 400)
        return
    }
    defer file.Close()
    // Process file
}
```

## Response Formatting

### Success Response
```go
func Handler(ctx *core.APIContext) {
    ctx.Success(map[string]interface{}{
        "message": "Operation completed",
    }, 200)
}
```

### Error Response
```go
func Handler(ctx *core.APIContext) {
    ctx.Error("Invalid input", 400)
}
```

### Custom Headers
```go
func Handler(ctx *core.APIContext) {
    ctx.Writer.Header().Set("X-Custom-Header", "value")
    ctx.Writer.Header().Set("Cache-Control", "no-cache")
    
    ctx.Success(map[string]string{
        "message": "Response with custom headers",
    }, 200)
}
```

## Best Practices

1. **Route Organization**
   - Group related routes in directories under app/api/
   - Use meaningful package names
   - Follow the init() pattern for registration
   - Keep handlers focused on a single responsibility

2. **Error Handling**
   - Validate input with appropriate error messages
   - Return clear errors with proper status codes
   - Use ctx.Error() for standardized error responses

3. **Security**
   - Validate all input parameters and request bodies
   - Sanitize output to prevent XSS and injection attacks
   - Use HTTPS in production

4. **Performance**
   - Minimize response size
   - Use appropriate caching headers
   - Handle timeouts properly

## Common Tasks

### Pagination
```go
func Handler(ctx *core.APIContext) {
    query := ctx.Request.URL.Query()
    page, _ := strconv.Atoi(query.Get("page"))
    limit, _ := strconv.Atoi(query.Get("limit"))
    
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20 // Default limit
    }
    
    offset := (page - 1) * limit
    // Fetch paginated data
}
```

### Search
```go
func Handler(ctx *core.APIContext) {
    query := ctx.Request.URL.Query().Get("q")
    // Implement search logic
}
```

### Filtering
```go
func Handler(ctx *core.APIContext) {
    filters := ctx.QueryParams() // Helper method to get all query params
    // Apply filters to data
}
```

## Troubleshooting

1. **Route Not Found**
   - Ensure you've registered the route with RegisterAPIHandler
   - Verify the API route path matches exactly
   - Check that the package is imported in main.go
   - Verify parameter format in dynamic routes [param]

2. **Request Errors**
   - Validate input format with appropriate type checking
   - Provide clear error messages for missing required fields
   - Check for nil values before accessing nested objects

3. **Response Issues**
   - Use ctx.Success() and ctx.Error() for standardized responses
   - Set appropriate HTTP status codes
   - Check content encoding when working with binary data 