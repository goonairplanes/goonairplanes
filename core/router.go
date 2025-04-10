package core

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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

func (r *Router) loadAPIRoutes() (int, error) {
	apiBasePath := filepath.Join(AppConfig.AppDir, "api")
	apiRouteCount := 0

	if _, err := os.Stat(apiBasePath); os.IsNotExist(err) {
		return 0, nil
	}

	err := filepath.Walk(apiBasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Base(path) == "route.go" {
			relPath, err := filepath.Rel(apiBasePath, filepath.Dir(path))
			if err != nil {
				return fmt.Errorf("failed to get relative path for %s: %w", path, err)
			}

			routePath := "/api/" + filepath.ToSlash(relPath)

			r.Routes = append(r.Routes, Route{
				Path: routePath,
				Handler: func(w http.ResponseWriter, req *http.Request) {
					if routePath == "/api/test" {
						handleAPITest(w, req)
					} else {
						r.Logger.WarnLog.Printf("API route %s not implemented", routePath)
						http.Error(w, "API route not implemented", http.StatusNotImplemented)
					}
				},
				IsAPI:      true,
				Middleware: NewMiddlewareChain(),
			})

			r.Logger.InfoLog.Printf("API route registered: %s", routePath)
			apiRouteCount++
		}

		return nil
	})

	return apiRouteCount, err
}

func handleAPITest(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Hello from Go on Airplanes API route!"}`))
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
