package api

import (
	"goonairplanes/core"
	"net/http"
)


type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}


var users = []User{
	{ID: 1, Name: "John Doe", Email: "john@example.com", Username: "johndoe"},
	{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Username: "janesmith"},
	{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", Username: "bobjohnson"},
}


func RegisterRoutes(router *core.Router) {
	
	router.API("/api/users", getUsers)

	
	router.API("/api/users/[id]", getUserByID)

	
	router.API("/api/users", createUser)
}


func getUsers(ctx *core.APIContext) {
	
	page, perPage := core.GetPaginationParams(ctx.Request, 10)

	
	totalItems := len(users)
	startIndex := (page - 1) * perPage
	endIndex := startIndex + perPage

	if startIndex >= totalItems {
		
		meta := core.NewPaginationMeta(page, perPage, totalItems)
		core.RenderPaginated(ctx.Writer, []User{}, meta, http.StatusOK)
		return
	}

	if endIndex > totalItems {
		endIndex = totalItems
	}

	
	pagedUsers := users[startIndex:endIndex]

	
	meta := core.NewPaginationMeta(page, perPage, totalItems)

	
	core.RenderPaginated(ctx.Writer, pagedUsers, meta, http.StatusOK)
}


func getUserByID(ctx *core.APIContext) {
	
	idStr := ctx.Params["id"]
	id := core.GetParamInt(ctx.Request, "id", 0)

	
	if id == 0 && idStr != "" {
		for i := range users {
			if users[i].ID == id {
				ctx.Success(users[i], http.StatusOK)
				return
			}
		}
	}

	
	ctx.Error("User not found", http.StatusNotFound)
}


func createUser(ctx *core.APIContext) {
	
	if ctx.Request.Method != http.MethodPost {
		ctx.Error("Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	
	var newUser User
	if err := ctx.ParseBody(&newUser); err != nil {
		ctx.Error("Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	
	if newUser.Name == "" || newUser.Email == "" || newUser.Username == "" {
		ctx.Error("Name, email and username are required", http.StatusBadRequest)
		return
	}

	
	newUser.ID = len(users) + 1

	
	users = append(users, newUser)

	
	ctx.Success(newUser, http.StatusCreated)
}
