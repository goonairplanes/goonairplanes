# JSON Utilities and API Helpers

Go on Airplanes includes a set of Ruby on Rails-inspired JSON utilities and API helpers to simplify the development of RESTful APIs.

## Core Features

- JSON response rendering with consistent format
- Request body parsing
- Query parameter extraction
- Pagination support
- API context with helper methods

## Usage Examples

### Basic API Route

```go
router.API("/api/products", func(ctx *core.APIContext) {
    // Get products from database
    products := []Product{
        {ID: 1, Name: "Product 1", Price: 9.99},
        {ID: 2, Name: "Product 2", Price: 19.99},
    }

    // Return JSON response
    ctx.Success(products, http.StatusOK)
})
```

### Parsing Request Body

```go
router.API("/api/products", func(ctx *core.APIContext) {
    // Only handle POST requests
    if ctx.Request.Method != http.MethodPost {
        ctx.Error("Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse request body
    var product Product
    if err := ctx.ParseBody(&product); err != nil {
        ctx.Error("Invalid request body: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Save product to database
    // ...

    // Return created product
    ctx.Success(product, http.StatusCreated)
})
```

### Handling URL Parameters

```go
router.API("/api/products/[id]", func(ctx *core.APIContext) {
    // Get product ID from path parameter
    productID := ctx.Params["id"]

    // Find product by ID
    // ...

    // Return product
    ctx.Success(product, http.StatusOK)
})
```

### Pagination Support

```go
router.API("/api/products", func(ctx *core.APIContext) {
    // Get pagination parameters (default 10 per page)
    page, perPage := core.GetPaginationParams(ctx.Request, 10)

    // Get paginated products
    totalItems := 100
    products := getProducts(page, perPage)

    // Create pagination metadata
    meta := core.NewPaginationMeta(page, perPage, totalItems)

    // Render paginated response
    core.RenderPaginated(ctx.Writer, products, meta, http.StatusOK)
})
```

## API Reference

### Response Functions

#### `RenderJSON(w http.ResponseWriter, data interface{}, statusCode int) error`
Renders data as JSON with the specified status code.

#### `RenderSuccess(w http.ResponseWriter, data interface{}, statusCode int) error`
Renders a success response with the specified data and status code.

#### `RenderError(w http.ResponseWriter, errMessage string, statusCode int) error`
Renders an error response with the specified message and status code.

#### `RenderPaginated(w http.ResponseWriter, data interface{}, meta PaginationMeta, statusCode int) error`
Renders a paginated response with the specified data, metadata, and status code.

### Request Parsing

#### `ParseBody(r *http.Request, v interface{}) error`
Parses the request body as JSON into the provided interface.

#### `ParseJSONParams(r *http.Request) map[string]interface{}`
Parses URL query parameters into a map.

#### `GetParam(r *http.Request, name string, defaultValue string) string`
Gets a query parameter with the specified name, returning the default value if not present.

#### `GetParamInt(r *http.Request, name string, defaultValue int) int`
Gets an integer query parameter with the specified name, returning the default value if not present or invalid.

#### `GetParamBool(r *http.Request, name string, defaultValue bool) bool`
Gets a boolean query parameter with the specified name, returning the default value if not present or invalid.

### Pagination

#### `NewPaginationMeta(currentPage, perPage, totalItems int) PaginationMeta`
Creates pagination metadata for the response.

#### `GetPaginationParams(r *http.Request, defaultPerPage int) (page, perPage int)`
Extracts pagination parameters from the request.

### Helper Methods

#### `IsJSONRequest(r *http.Request) bool`
Checks if the request content type is JSON.

#### `APIResponse(success bool, data interface{}, errMessage string) ResponseData`
Creates a standard API response struct.

## APIContext Methods

The `APIContext` type provides convenience methods for API routes:

- `Success(data interface{}, statusCode int)` - Renders a success response
- `Error(message string, statusCode int)` - Renders an error response
- `ParseBody(v interface{}) error` - Parses the request body
- `QueryParams() map[string]interface{}` - Gets all query parameters 