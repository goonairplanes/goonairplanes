package api

import (
	"goonairplanes/core"
	"net/http"
)

// User represents a sample user structure
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// Mock data for demonstration
var users = []User{
	{ID: 1, Name: "John Doe", Email: "john@example.com", Username: "johndoe"},
	{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Username: "janesmith"},
	{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", Username: "bobjohnson"},
}

// RegisterRoutes registers all routes for the users API
func RegisterRoutes(router *core.Router) {
	// Get all users with pagination
	router.API("/api/users", getUsers)

	// Get a specific user by ID
	router.API("/api/users/[id]", getUserByID)

	// Create a new user
	router.API("/api/users", createUser)
}

// getUsers handles GET /api/users
func getUsers(ctx *core.APIContext) {
	// Extract pagination parameters
	page, perPage := core.GetPaginationParams(ctx.Request, 10)

	// Calculate pagination
	totalItems := len(users)
	startIndex := (page - 1) * perPage
	endIndex := startIndex + perPage

	if startIndex >= totalItems {
		// If page is out of bounds, return empty array with pagination info
		meta := core.NewPaginationMeta(page, perPage, totalItems)
		core.RenderPaginated(ctx.Writer, []User{}, meta, http.StatusOK)
		return
	}

	if endIndex > totalItems {
		endIndex = totalItems
	}

	// Get paginated results
	pagedUsers := users[startIndex:endIndex]

	// Create pagination metadata
	meta := core.NewPaginationMeta(page, perPage, totalItems)

	// Render response with pagination
	core.RenderPaginated(ctx.Writer, pagedUsers, meta, http.StatusOK)
}

// getUserByID handles GET /api/users/[id]
func getUserByID(ctx *core.APIContext) {
	// Get user ID from params
	idStr := ctx.Params["id"]
	id := core.GetParamInt(ctx.Request, "id", 0)

	// If ID was in the path parameters and not in query
	if id == 0 && idStr != "" {
		for i := range users {
			if users[i].ID == id {
				ctx.Success(users[i], http.StatusOK)
				return
			}
		}
	}

	// User not found
	ctx.Error("User not found", http.StatusNotFound)
}

// createUser handles POST /api/users
func createUser(ctx *core.APIContext) {
	// Only allow POST method
	if ctx.Request.Method != http.MethodPost {
		ctx.Error("Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var newUser User
	if err := ctx.ParseBody(&newUser); err != nil {
		ctx.Error("Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate user data
	if newUser.Name == "" || newUser.Email == "" || newUser.Username == "" {
		ctx.Error("Name, email and username are required", http.StatusBadRequest)
		return
	}

	// Generate new ID (in a real app, this would be handled by the database)
	newUser.ID = len(users) + 1

	// Add to users slice (in a real app, this would be saved to a database)
	users = append(users, newUser)

	// Return the created user
	ctx.Success(newUser, http.StatusCreated)
}
