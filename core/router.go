package core

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var paramRegex = regexp.MustCompile(`\[([^/\]]+)\]`)

type Route struct {
	Path       string
	Handler    http.HandlerFunc
	ParamNames []string
	IsStatic   bool
	IsAPI      bool
	IsParam    bool
	Pattern    *regexp.Regexp
	Middleware *MiddlewareChain
}

type Router struct {
	Routes           []Route
	Marley           *Marley
	StaticDir        string
	Logger           *AppLogger
	GlobalMiddleware *MiddlewareChain
}

type RouteContext struct {
	Params map[string]string
	Config *Config
}

type APIHandler interface {
	Handler(w http.ResponseWriter, r *http.Request)
}

type APIContext struct {
	Request *http.Request
	Writer  http.ResponseWriter
	Params  map[string]string
	Config  *Config
}

func (ctx *APIContext) Success(data interface{}, statusCode int) {
	RenderSuccess(ctx.Writer, data, statusCode)
}

func (ctx *APIContext) Error(message string, statusCode int) {
	RenderError(ctx.Writer, message, statusCode)
}

func (ctx *APIContext) ParseBody(v interface{}) error {
	return ParseBody(ctx.Request, v)
}

func (ctx *APIContext) QueryParams() map[string]interface{} {
	return ParseJSONParams(ctx.Request)
}

func NewRouter(logger *AppLogger) *Router {
	return &Router{
		Routes:           []Route{},
		Marley:           NewMarley(logger),
		StaticDir:        AppConfig.StaticDir,
		Logger:           logger,
		GlobalMiddleware: NewMiddlewareChain(),
	}
}

func (r *Router) Use(middleware MiddlewareFunc) {
	r.GlobalMiddleware.Use(middleware)
}

func (r *Router) AddRoute(path string, handler http.HandlerFunc, middleware ...MiddlewareFunc) {
	mc := NewMiddlewareChain()
	for _, m := range middleware {
		mc.Use(m)
	}

	paramNames := r.extractParamNames(path)
	isParam := len(paramNames) > 0

	var pattern *regexp.Regexp
	if isParam {
		patternStr := "^" + paramRegex.ReplaceAllString(path, "([^/]+)") + "$"
		pattern = regexp.MustCompile(patternStr)
	}

	r.Routes = append(r.Routes, Route{
		Path:       path,
		Handler:    handler,
		ParamNames: paramNames,
		IsStatic:   false,
		IsAPI:      false,
		IsParam:    isParam,
		Pattern:    pattern,
		Middleware: mc,
	})
}

func (r *Router) AddAPIRoute(path string, handler http.HandlerFunc, middleware ...MiddlewareFunc) {
	mc := NewMiddlewareChain()
	for _, m := range middleware {
		mc.Use(m)
	}

	r.Routes = append(r.Routes, Route{
		Path:       path,
		Handler:    handler,
		IsAPI:      true,
		Middleware: mc,
	})
}

func (r *Router) API(path string, handler func(*APIContext), middleware ...MiddlewareFunc) {
	wrappedHandler := func(w http.ResponseWriter, req *http.Request) {
		params := extractParamsFromRequest(req.URL.Path, path)
		ctx := &APIContext{
			Request: req,
			Writer:  w,
			Params:  params,
			Config:  &AppConfig,
		}
		handler(ctx)
	}

	r.AddAPIRoute(path, wrappedHandler, middleware...)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if AppConfig.LogLevel != "error" {
		r.Logger.InfoLog.Printf("%s %s", req.Method, req.URL.Path)
	}

	path := normalizePath(req.URL.Path)

	if r.GlobalMiddleware == nil {
		r.GlobalMiddleware = NewMiddlewareChain()
	}

	if strings.HasPrefix(path, "/static") {
		for _, route := range r.Routes {
			if route.IsStatic {
				if route.Middleware == nil {
					route.Middleware = NewMiddlewareChain()
				}
				handler := r.GlobalMiddleware.Then(http.HandlerFunc(route.Handler))
				handler.ServeHTTP(w, req)
				return
			}
		}
	}

	if strings.HasPrefix(path, "/api") {
		for _, route := range r.Routes {
			if route.IsAPI {
				apiPath := route.Path
				if path == apiPath || strings.HasPrefix(path, apiPath+"/") {
					if route.Middleware == nil {
						route.Middleware = NewMiddlewareChain()
					}
					handler := r.GlobalMiddleware.Then(route.Middleware.Then(http.HandlerFunc(route.Handler)))
					handler.ServeHTTP(w, req)
					return
				}
			}
		}

		r.Logger.WarnLog.Printf("API route not found: %s", path)
		http.Error(w, "API endpoint not found", http.StatusNotFound)
		return
	}

	for _, route := range r.Routes {
		if !route.IsParam && !route.IsStatic && !route.IsAPI {
			routePath := route.Path
			if path == routePath {
				if route.Middleware == nil {
					route.Middleware = NewMiddlewareChain()
				}
				handler := r.GlobalMiddleware.Then(route.Middleware.Then(http.HandlerFunc(route.Handler)))
				handler.ServeHTTP(w, req)
				return
			}
		}
	}

	for _, route := range r.Routes {
		if route.IsParam && route.Pattern != nil {
			if route.Pattern.MatchString(path) {
				if route.Middleware == nil {
					route.Middleware = NewMiddlewareChain()
				}
				handler := r.GlobalMiddleware.Then(route.Middleware.Then(http.HandlerFunc(route.Handler)))
				handler.ServeHTTP(w, req)
				return
			}
		}
	}

	r.Logger.WarnLog.Printf("Route not found: %s", path)
	r.serveErrorPage(w, req, http.StatusNotFound)
}

func (r *Router) InitRoutes() error {
	startTime := time.Now()
	r.Logger.InfoLog.Printf("Initializing routes...")

	r.Routes = []Route{}

	err := r.Marley.LoadTemplates()
	if err != nil {
		r.Logger.ErrorLog.Printf("Failed to load templates: %v", err)
		return fmt.Errorf("failed to load templates: %w", err)
	}

	r.AddStaticRoute()

	routeCount := 0
	for routePath := range r.Marley.Templates {
		paramNames := r.extractParamNames(routePath)
		isParam := len(paramNames) > 0

		var pattern *regexp.Regexp
		if isParam {
			patternStr := "^" + paramRegex.ReplaceAllString(routePath, "([^/]+)") + "$"
			pattern = regexp.MustCompile(patternStr)
		}

		r.Routes = append(r.Routes, Route{
			Path:       routePath,
			Handler:    r.createTemplateHandler(routePath),
			ParamNames: paramNames,
			IsStatic:   false,
			IsAPI:      false,
			IsParam:    isParam,
			Pattern:    pattern,
			Middleware: NewMiddlewareChain(),
		})

		r.Logger.InfoLog.Printf("Route registered: %s (params: %v)", routePath, paramNames)
		routeCount++
	}

	apiRouteCount, err := r.loadAPIRoutes()
	if err != nil {
		r.Logger.ErrorLog.Printf("Failed to load API routes: %v", err)
		return fmt.Errorf("failed to load API routes: %w", err)
	}

	elapsedTime := time.Since(startTime)
	r.Logger.InfoLog.Printf("Routes initialized: %d page routes, %d API routes in %v",
		routeCount, apiRouteCount, elapsedTime.Round(time.Millisecond))

	return nil
}

func (r *Router) loadAPIRoutes() (int, error) {
	apiBasePath := filepath.Join(AppConfig.AppDir, "api")
	apiRouteCount := 0

	if _, err := os.Stat(apiBasePath); os.IsNotExist(err) {
		return 0, nil
	}

	
	usersPath := filepath.Join(apiBasePath, "users")
	if _, err := os.Stat(usersPath); err == nil {
		
		r.API("/api/users", func(ctx *APIContext) {
			
			switch ctx.Request.Method {
			case http.MethodGet:
				
				r.handleUsersGet(ctx)
			case http.MethodPost:
				
				r.handleUsersPost(ctx)
			default:
				ctx.Error("Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		
		r.API("/api/users/[id]", func(ctx *APIContext) {
			
			switch ctx.Request.Method {
			case http.MethodGet:
				
				r.handleUserGetById(ctx)
			case http.MethodPut:
				
				r.handleUserPutById(ctx)
			case http.MethodDelete:
				
				r.handleUserDeleteById(ctx)
			default:
				ctx.Error("Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		r.Logger.InfoLog.Printf("API route registered: %s", "/api/users")
		apiRouteCount += 2 
	}

	
	testPath := filepath.Join(apiBasePath, "test")
	if _, err := os.Stat(testPath); err == nil {
		
		r.API("/api/test", func(ctx *APIContext) {
			
			r.handleTestAPI(ctx)
		})

		r.Logger.InfoLog.Printf("API route registered: %s", "/api/test")
		apiRouteCount++
	}

	return apiRouteCount, nil
}


var users = []map[string]interface{}{
	{"id": 1, "name": "John Doe", "email": "john@example.com", "username": "johndoe"},
	{"id": 2, "name": "Jane Smith", "email": "jane@example.com", "username": "janesmith"},
	{"id": 3, "name": "Bob Johnson", "email": "bob@example.com", "username": "bobjohnson"},
}


func (r *Router) handleUsersGet(ctx *APIContext) {
	
	page, perPage := GetPaginationParams(ctx.Request, 10)

	
	totalItems := len(users)
	startIndex := (page - 1) * perPage
	endIndex := startIndex + perPage

	if startIndex >= totalItems {
		
		meta := NewPaginationMeta(page, perPage, totalItems)
		RenderPaginated(ctx.Writer, []map[string]interface{}{}, meta, http.StatusOK)
		return
	}

	if endIndex > totalItems {
		endIndex = totalItems
	}

	
	pagedUsers := users[startIndex:endIndex]

	
	meta := NewPaginationMeta(page, perPage, totalItems)

	
	RenderPaginated(ctx.Writer, pagedUsers, meta, http.StatusOK)
}


func (r *Router) handleUsersPost(ctx *APIContext) {
	
	var newUser map[string]interface{}
	if err := ParseBody(ctx.Request, &newUser); err != nil {
		ctx.Error("Invalid request body", http.StatusBadRequest)
		return
	}

	
	if newUser["name"] == nil || newUser["email"] == nil || newUser["username"] == nil {
		ctx.Error("Name, email and username are required", http.StatusBadRequest)
		return
	}

	
	newUser["id"] = len(users) + 1

	
	users = append(users, newUser)

	
	ctx.Success(newUser, http.StatusCreated)
}


func (r *Router) handleUserGetById(ctx *APIContext) {
	
	idStr := ctx.Params["id"]
	id := 0
	if idStr != "" {
		id, _ = strconv.Atoi(idStr)
	}

	if id <= 0 {
		ctx.Error("Invalid user ID", http.StatusBadRequest)
		return
	}

	
	for _, user := range users {
		if userId, ok := user["id"].(int); ok && userId == id {
			ctx.Success(user, http.StatusOK)
			return
		}
	}

	
	ctx.Error("User not found", http.StatusNotFound)
}


func (r *Router) handleUserPutById(ctx *APIContext) {
	
	idStr := ctx.Params["id"]
	id := 0
	if idStr != "" {
		id, _ = strconv.Atoi(idStr)
	}

	if id <= 0 {
		ctx.Error("Invalid user ID", http.StatusBadRequest)
		return
	}

	
	var updatedUser map[string]interface{}
	if err := ParseBody(ctx.Request, &updatedUser); err != nil {
		ctx.Error("Invalid request body", http.StatusBadRequest)
		return
	}

	
	for i, user := range users {
		if userId, ok := user["id"].(int); ok && userId == id {
			
			updatedUser["id"] = id

			
			users[i] = updatedUser

			
			ctx.Success(updatedUser, http.StatusOK)
			return
		}
	}

	
	ctx.Error("User not found", http.StatusNotFound)
}


func (r *Router) handleUserDeleteById(ctx *APIContext) {
	
	idStr := ctx.Params["id"]
	id := 0
	if idStr != "" {
		id, _ = strconv.Atoi(idStr)
	}

	if id <= 0 {
		ctx.Error("Invalid user ID", http.StatusBadRequest)
		return
	}

	
	for i, user := range users {
		if userId, ok := user["id"].(int); ok && userId == id {
			
			users = append(users[:i], users[i+1:]...)

			
			ctx.Success(nil, http.StatusNoContent)
			return
		}
	}

	
	ctx.Error("User not found", http.StatusNotFound)
}


func (r *Router) handleTestAPI(ctx *APIContext) {
	
	response := map[string]interface{}{
		"message":   "Hello from Go on Airplanes API route!",
		"timestamp": time.Now().Format(time.RFC3339),
		"method":    ctx.Request.Method,
		"path":      ctx.Request.URL.Path,
		"params":    ctx.Params,
	}

	
	ctx.Success(response, http.StatusOK)
}

func (r *Router) AddStaticRoute() {
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(r.StaticDir)))
	r.Routes = append(r.Routes, Route{
		Path: "/static/",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			staticHandler.ServeHTTP(w, req)
		},
		IsStatic:   true,
		Middleware: NewMiddlewareChain(),
	})

	r.Logger.InfoLog.Printf("Static route registered: /static/ → %s", r.StaticDir)
}

func (r *Router) createTemplateHandler(route string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		startTime := time.Now()

		params := extractParamsFromRequest(req.URL.Path, route)
		ctx := &RouteContext{
			Params: params,
			Config: &AppConfig,
		}

		err := r.Marley.RenderTemplate(w, route, ctx)
		if err != nil {
			r.Logger.ErrorLog.Printf("Template rendering error for %s: %v", route, err)
			r.serveErrorPage(w, req, http.StatusInternalServerError)
			return
		}

		if AppConfig.LogLevel == "debug" {
			elapsedTime := time.Since(startTime)
			r.Logger.InfoLog.Printf("Rendered %s in %v", route, elapsedTime.Round(time.Microsecond))
		}
	}
}

func (r *Router) serveErrorPage(w http.ResponseWriter, req *http.Request, status int) {
	var errorPage string

	switch status {
	case http.StatusNotFound:
		errorPage = "404"
	case http.StatusInternalServerError:
		errorPage = "500"
	default:
		errorPage = "error"
	}

	customErrorPath := filepath.Join(AppConfig.AppDir, errorPage+".html")
	if _, err := os.Stat(customErrorPath); err == nil {
		ctx := &RouteContext{
			Params: map[string]string{
				"status": fmt.Sprintf("%d", status),
				"path":   req.URL.Path,
			},
			Config: &AppConfig,
		}

		if tmpl, exists := r.Marley.Templates["/"+errorPage]; exists {
			w.WriteHeader(status)
			if err := tmpl.ExecuteTemplate(w, "layout", ctx); err == nil {
				return
			}
		}
	}

	http.Error(w, http.StatusText(status), status)
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}

	if path != "/" && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	return path
}

func (r *Router) extractParamNames(routePath string) []string {
	matches := paramRegex.FindAllStringSubmatch(routePath, -1)
	var paramNames []string

	for _, match := range matches {
		if len(match) > 1 {
			paramNames = append(paramNames, match[1])
		}
	}

	return paramNames
}

func extractParamsFromRequest(requestPath, routePath string) map[string]string {
	params := make(map[string]string)

	requestPath = normalizePath(requestPath)

	if !strings.Contains(routePath, "[") {
		return params
	}

	patternStr := "^" + paramRegex.ReplaceAllString(routePath, "([^/]+)") + "$"
	pattern := regexp.MustCompile(patternStr)

	matches := pattern.FindStringSubmatch(requestPath)
	if len(matches) <= 1 {
		return params
	}

	paramNames := paramRegex.FindAllStringSubmatch(routePath, -1)

	for i, match := range paramNames {
		if i+1 < len(matches) && len(match) > 1 {
			params[match[1]] = matches[i+1]
		}
	}

	return params
}
