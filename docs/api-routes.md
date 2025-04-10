# 🔌 API Routes

This guide covers how to create and work with API routes in your GOA application.

## Creating Routes

### Basic Route
```go
// app/api/hello/route.go
func Handler(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Hello!",
    })
}
```

### Route with Parameters
```go
// app/api/user/route.go
func Handler(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    json.NewEncoder(w).Encode(map[string]string{
        "id": userID,
        "name": "John Doe",
    })
}
```

### POST Route
```go
// app/api/create/route.go
func Handler(w http.ResponseWriter, r *http.Request) {
    var data map[string]interface{}
    json.NewDecoder(r.Body).Decode(&data)
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data": data,
    })
}
```

## Request Handling

### Reading Query Parameters
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    page := r.URL.Query().Get("page")
    limit := r.URL.Query().Get("limit")
    // Use parameters
}
```

### Reading POST Data
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }
    json.NewDecoder(r.Body).Decode(&input)
    // Use input data
}
```

### File Upload
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()
    // Process file
}
```

## Response Formatting

### Success Response
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data": map[string]string{
            "message": "Operation completed",
        },
    })
}
```

### Error Response
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "error",
        "message": "Invalid input",
    })
}
```

### Custom Headers
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("X-Custom-Header", "value")
    w.Header().Set("Cache-Control", "no-cache")
    // Response
}
```

## Best Practices

1. **Route Organization**
   - Group related routes
   - Use meaningful names
   - Keep handlers focused

2. **Error Handling**
   - Validate input
   - Return clear errors
   - Use appropriate status codes

3. **Security**
   - Validate all input
   - Sanitize output
   - Use HTTPS

4. **Performance**
   - Minimize response size
   - Use caching
   - Handle timeouts

## Common Tasks

### Pagination
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    
    offset := (page - 1) * limit
    // Fetch paginated data
}
```

### Search
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    // Implement search logic
}
```

### Filtering
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    filters := r.URL.Query()
    // Apply filters to data
}
```

## Troubleshooting

1. **Route Not Found**
   - Check route registration
   - Verify URL pattern
   - Check middleware

2. **Request Errors**
   - Validate input format
   - Check required fields
   - Handle missing data

3. **Response Issues**
   - Set correct content type
   - Check encoding
   - Verify status codes 