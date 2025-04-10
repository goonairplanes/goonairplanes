# 🚀 Go on Airplanes Documentation

Welcome to the Go on Airplanes framework documentation! This guide focuses on what you need to know to build your application.

## 📋 Quick Start

1. [Installation](getting-started.md#installation)
2. [Your First Page](getting-started.md#your-first-page)
3. [Adding Components](getting-started.md#adding-components)
4. [Creating API Routes](getting-started.md#creating-api-routes)

## 🛠️ Core Features

### Templates
- [Basic Templates](templates.md#basic-templates)
- [Layouts](templates.md#layouts)
- [Components](templates.md#components)
- [Template Functions](templates.md#template-functions)

### Middleware
- [Using Middleware](middleware.md#using-middleware)
- [Built-in Middleware](middleware.md#built-in-middleware)
- [Custom Middleware](middleware.md#custom-middleware)

### API Routes
- [Creating Routes](api-routes.md#creating-routes)
- [Request Handling](api-routes.md#request-handling)
- [Response Formatting](api-routes.md#response-formatting)
- [JSON Utilities](JSON_UTILS.md#core-features)
- [Pagination Support](JSON_UTILS.md#pagination-support)
- [API Context Helpers](JSON_UTILS.md#apicontext-methods)

### Client-Side
- [jQuery Integration](client-side.md#jquery-integration)
- [Dynamic Content](client-side.md#dynamic-content)
- [Form Handling](client-side.md#form-handling)

## 📚 Examples

### Basic Website
```go
// app/index.html
{{define "content"}}
<h1>Welcome</h1>
{{template "button" (dict "Text" "Click Me")}}
{{end}}
```

### API Endpoint
```go
// app/api/hello/route.go
func Handler(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Hello!",
    })
}
```

### Middleware
```go
// app/middleware.go
func ConfigureMiddleware(app *core.GonAirApp) {
    app.Router.Use(core.LoggingMiddleware(app.Logger))
    app.Router.Use(core.AuthMiddleware(validateToken))
}
```

## 🎯 Best Practices

1. **Templates**
   - Keep templates small and focused
   - Use components for reusable UI
   - Follow consistent naming

2. **Middleware**
   - Add critical middleware first
   - Keep middleware simple
   - Handle errors properly

3. **API Routes**
   - Use meaningful route names
   - Return consistent response formats
   - Handle errors gracefully
   - Use JSON utilities for standardized responses
   - Implement proper pagination for list endpoints
   - Validate request bodies and parameters

4. **Client-Side**
   - Use jQuery for DOM manipulation
   - Keep JavaScript organized
   - Handle errors in AJAX calls

## ❓ Need Help?

- [GitHub Issues](https://github.com/yourusername/goonairplanes/issues)
- [Community Discussions](https://github.com/yourusername/goonairplanes/discussions) 