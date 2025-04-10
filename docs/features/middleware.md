# 🔒 Middleware System

The middleware system in Go on Airplanes provides a flexible way to handle HTTP requests and responses. Middleware functions can be used for authentication, logging, error handling, and more.

## 📋 Table of Contents

- [Overview](#overview)
- [Built-in Middleware](#built-in-middleware)
- [Configuration](#configuration)
- [Creating Custom Middleware](#creating-custom-middleware)
- [Examples](#examples)

## Overview

Middleware in Go on Airplanes is implemented as a chain of functions that can:
- Process requests before they reach your handlers
- Process responses after they leave your handlers
- Modify both requests and responses
- Stop the request chain if needed

## Built-in Middleware

The framework includes several built-in middleware options:

### Logging Middleware
```go
app.Router.Use(core.LoggingMiddleware(app.Logger))
```
Records request details including method, path, and timing.

### Recovery Middleware
```go
app.Router.Use(core.RecoveryMiddleware(app.Logger))
```
Gracefully handles panics during request processing.

### Authentication Middleware
```go
app.Router.Use(core.AuthMiddleware(func(token string) bool {
    // Implement your token validation logic
    return true
}))
```
Protects routes with token-based authentication.

### CORS Middleware
```go
app.Router.Use(core.CORSMiddleware([]string{"*"}))
```
Configures Cross-Origin Resource Sharing.

### Rate Limiting Middleware
```go
app.Router.Use(core.RateLimitMiddleware(100))
```
Limits request frequency to prevent abuse.

### Secure Headers Middleware
```go
app.Router.Use(core.SecureHeadersMiddleware())
```
Adds security-related HTTP headers.

## Configuration

Configure middleware in `app/middleware.go`:

```go
package app

import (
    "goonairplanes/core"
)

func ConfigureMiddleware(app *core.GonAirApp) {
    // Global middleware
    app.Router.Use(core.LoggingMiddleware(app.Logger))
    app.Router.Use(core.RecoveryMiddleware(app.Logger))
    
    // Conditional middleware
    if app.Config.EnableCORS {
        app.Router.Use(core.CORSMiddleware(app.Config.AllowedOrigins))
    }
    
    // Route-specific middleware
    app.Router.AddRoute("/dashboard", handler, 
        core.AuthMiddleware(validateToken))
}
```

## Creating Custom Middleware

Create your own middleware functions:

```go
func CustomMiddleware(app *core.GonAirApp) core.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Pre-processing
            app.Logger.InfoLog.Printf("Processing request: %s", r.URL.Path)
            
            // Call next handler
            next.ServeHTTP(w, r)
            
            // Post-processing
        })
    }
}
```

## Examples

### Basic Authentication
```go
func AuthMiddleware(validateToken func(string) bool) core.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if token == "" || !validateToken(token) {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### Request Timing
```go
func TimingMiddleware(app *core.GonAirApp) core.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            next.ServeHTTP(w, r)
            duration := time.Since(start)
            app.Logger.InfoLog.Printf("Request took %v", duration)
        })
    }
}
```

### Error Handling
```go
func ErrorHandlingMiddleware(app *core.GonAirApp) core.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    app.Logger.ErrorLog.Printf("Panic recovered: %v", err)
                    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

## Best Practices

1. **Order Matters**: Middleware is executed in the order it's added. Add critical middleware first.
2. **Keep it Simple**: Each middleware should do one thing well.
3. **Error Handling**: Always handle errors gracefully.
4. **Performance**: Be mindful of middleware that adds significant overhead.
5. **Logging**: Use appropriate log levels for different types of information.

## Troubleshooting

Common issues and solutions:

1. **Middleware Not Executing**
   - Check the order of middleware registration
   - Verify the route pattern matches
   - Ensure middleware is properly chained

2. **Performance Issues**
   - Profile middleware execution time
   - Consider caching expensive operations
   - Remove unnecessary middleware in production

3. **Authentication Problems**
   - Verify token validation logic
   - Check header names and formats
   - Ensure proper error responses 