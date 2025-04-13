package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var paramRegex = regexp.MustCompile(`\[([^/\]]+)\]`)

var apiRegistry = make(map[string]map[string]func(*APIContext))
var apiRegistryMutex sync.RWMutex

func RegisterAPIHandler(path string, method string, handler func(*APIContext)) {
	apiRegistryMutex.Lock()
	defer apiRegistryMutex.Unlock()

	path = normalizePath(path)

	if _, ok := apiRegistry[path]; !ok {
		apiRegistry[path] = make(map[string]func(*APIContext))
	}
	apiRegistry[path][method] = handler

}

type Route struct {
	Path       string
	Handler    http.HandlerFunc
	ParamNames []string
	IsStatic   bool
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
	mutex            sync.RWMutex
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
		IsParam:    isParam,
		Pattern:    pattern,
		Middleware: mc,
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()
	requestPath := normalizePath(req.URL.Path)

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(requestPath, "/static") {
			for _, route := range r.Routes {
				if route.IsStatic {
					route.Handler.ServeHTTP(w, req)
					return
				}
			}
		}

		if strings.HasPrefix(requestPath, "/api") {
			apiRegistryMutex.RLock()
			var matchedHandler func(*APIContext)
			var matchedParams map[string]string
			var matchedPath string

			for registeredPath, methodMap := range apiRegistry {
				params, ok := matchPath(registeredPath, requestPath)
				if ok {
					matchedPath = registeredPath
					if handler, methodExists := methodMap[req.Method]; methodExists {
						matchedHandler = handler
						matchedParams = params
					} else if handler, anyMethodExists := methodMap["*"]; anyMethodExists {
						matchedHandler = handler
						matchedParams = params
					}
				}
			}
			apiRegistryMutex.RUnlock()

			if HasAPIError(matchedPath, req.Method) {
				apiErr := GetAPIError(matchedPath, req.Method)
				r.Logger.WarnLog.Printf("API endpoint has known error: %s %s: %s",
					req.Method, matchedPath, apiErr.ErrorMsg)

				RenderError(w, apiErr.ErrorMsg, apiErr.Code)
				return
			}

			if matchedHandler != nil {

				ctx := &APIContext{
					Request: req,
					Writer:  w,
					Params:  matchedParams,
					Config:  &AppConfig,
				}

				func() {
					defer func() {
						if rec := recover(); rec != nil {
							errMsg := fmt.Sprintf("API handler panic: %v", rec)
							r.Logger.ErrorLog.Printf("%s %s - %s", req.Method, matchedPath, errMsg)

							RegisterAPIError(matchedPath, req.Method, fmt.Errorf("%v", rec), http.StatusInternalServerError)

							RenderError(w, "Internal Server Error", http.StatusInternalServerError)
						}
					}()

					matchedHandler(ctx)
				}()
				return
			}

			RenderError(w, "API endpoint not found or method not allowed", http.StatusNotFound)
			return
		}

		var pageHandler http.HandlerFunc
		var pageMiddleware *MiddlewareChain

		for _, route := range r.Routes {
			if !route.IsParam && !route.IsStatic {
				if requestPath == route.Path {
					pageHandler = route.Handler
					pageMiddleware = route.Middleware
					break
				}
			}
		}

		if pageHandler == nil {
			for _, route := range r.Routes {
				if route.IsParam && route.Pattern != nil {
					if route.Pattern.MatchString(requestPath) {
						pageHandler = route.Handler
						pageMiddleware = route.Middleware
						break
					}
				}
			}
		}

		if pageHandler != nil {
			if pageMiddleware == nil {
				pageMiddleware = NewMiddlewareChain()
			}
			finalPageRouteHandler := pageMiddleware.Then(pageHandler)
			finalPageRouteHandler.ServeHTTP(w, req)
			return
		}

		r.serveErrorPage(w, req, http.StatusNotFound)
	})

	if r.GlobalMiddleware != nil {
		r.GlobalMiddleware.Then(finalHandler).ServeHTTP(w, req)
	} else {
		finalHandler.ServeHTTP(w, req)
	}

	if AppConfig.LogLevel != "error" {
		go r.logRequest(req, http.StatusOK, time.Since(startTime))
	}
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

	pageRouteCount := 0
	for routePath := range r.Marley.Templates {
		if filepath.Base(routePath) == "layout.html" {
			continue
		}

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
			IsParam:    isParam,
			Pattern:    pattern,
			Middleware: NewMiddlewareChain(),
		})

		r.Logger.InfoLog.Printf("Page route registered: %s (params: %v)", routePath, paramNames)
		pageRouteCount++
	}

	apiRouteCount := r.discoverAndLogAPIRoutes()

	elapsedTime := time.Since(startTime)
	r.Logger.InfoLog.Printf("Routes initialized: %d page routes discovered, %d API routes discovered in %v. API handlers registered via init().",
		pageRouteCount, apiRouteCount, elapsedTime.Round(time.Millisecond))

	apiRegistryMutex.RLock()
	r.Logger.InfoLog.Printf("--- Registered API Handlers ---")
	for path, methodMap := range apiRegistry {
		for method := range methodMap {
			r.Logger.InfoLog.Printf("  %s %s", method, path)
		}
	}
	r.Logger.InfoLog.Printf("-----------------------------")
	apiRegistryMutex.RUnlock()

	return nil
}

func (r *Router) discoverAndLogAPIRoutes() int {
	apiBasePath := filepath.Join(AppConfig.AppDir, "api")
	discoveredCount := 0

	if _, err := os.Stat(apiBasePath); os.IsNotExist(err) {
		r.Logger.InfoLog.Printf("No 'api' directory found in '%s'. Skipping API route discovery.", AppConfig.AppDir)
		return 0
	}

	filepath.Walk(apiBasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			r.Logger.WarnLog.Printf("Error accessing path %q: %v", path, err)
			return err
		}

		if !info.IsDir() && info.Name() == "route.go" {
			relPath, err := filepath.Rel(apiBasePath, filepath.Dir(path))
			if err != nil {
				r.Logger.WarnLog.Printf("Could not get relative path for %s: %v", path, err)
				return nil
			}

			relPath = filepath.ToSlash(relPath)

			apiRoutePath := "/api"
			if relPath != "." {
				parts := strings.Split(relPath, "/")
				for _, part := range parts {
					if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
					}
				}
				apiRoutePath = "/api/" + strings.Join(parts, "/")
			}

			apiRoutePath = normalizePath(apiRoutePath)

			r.Logger.InfoLog.Printf("Discovered potential API route file for: %s", apiRoutePath)
			discoveredCount++
		}
		return nil
	})

	return discoveredCount
}

func (r *Router) AddStaticRoute() {
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(r.StaticDir)))
	r.Routes = append(r.Routes, Route{
		Path: "/static/",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			if _, err := os.Stat(r.StaticDir); os.IsNotExist(err) {
				r.Logger.ErrorLog.Printf("Static directory '%s' not found", r.StaticDir)
				http.NotFound(w, req)
				return
			}
			staticHandler.ServeHTTP(w, req)
		},
		IsStatic:   true,
		Middleware: NewMiddlewareChain(),
	})
	r.Logger.InfoLog.Printf("Static route registered: /static/ -> %s", r.StaticDir)
}

func (r *Router) createTemplateHandler(routePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestPath := normalizePath(req.URL.Path)
		params := extractParamsFromRequest(requestPath, routePath)

		data := map[string]interface{}{
			"Params":     params,
			"Config":     &AppConfig,
			"ServerTime": time.Now().Format(time.RFC1123),
			"BuildTime":  time.Now().Format(time.RFC1123),
			"Route":      routePath,
			"Request": map[string]interface{}{
				"Path":   requestPath,
				"Method": req.Method,
				"Host":   req.Host,
			},
		}

		if HasPageError(routePath) {
			r.Logger.WarnLog.Printf("Rendering error page for route with known errors: %s", routePath)
			r.RenderErrorPage(w, req, routePath)
			return
		}

		r.Marley.mutex.RLock()
		_, hasErrors := r.Marley.TemplateErrors[routePath]
		r.Marley.mutex.RUnlock()

		if hasErrors {
			r.Logger.WarnLog.Printf("Skipping render of template with known errors: %s", routePath)
			r.RenderErrorPage(w, req, routePath)
			return
		}

		err := r.Marley.RenderTemplate(w, routePath, data)
		if err != nil {
			r.Logger.ErrorLog.Printf("Template rendering error for request %s (template %s): %v", requestPath, routePath, err)
			statusCode := http.StatusInternalServerError

			if strings.Contains(err.Error(), "template not found") {
				statusCode = http.StatusNotFound
			}

			RegisterPageError(routePath, err, statusCode)

			r.RenderErrorPage(w, req, routePath)
			return
		}
	}
}

func (r *Router) serveErrorPage(w http.ResponseWriter, req *http.Request, status int, customMessage ...string) {
	var errorPage string
	var errorMessage string

	if len(customMessage) > 0 {
		errorMessage = customMessage[0]
	} else {
		errorMessage = http.StatusText(status)
	}

	switch status {
	case http.StatusNotFound:
		errorPage = "404"
	case http.StatusInternalServerError:
		errorPage = "500"
	default:
		errorPage = fmt.Sprintf("%d", status)
		if _, exists := r.Marley.Templates["/"+errorPage]; !exists {
			errorPage = "error"
		}
	}

	errorTemplatePath := "/" + errorPage

	r.Marley.mutex.RLock()
	_, hasErrors := r.Marley.TemplateErrors[errorTemplatePath]
	r.Marley.mutex.RUnlock()

	if tmpl, exists := r.Marley.Templates[errorTemplatePath]; exists && !hasErrors {
		w.WriteHeader(status)
		err := tmpl.ExecuteTemplate(w, "layout", map[string]interface{}{
			"Params": map[string]string{
				"status":  fmt.Sprintf("%d", status),
				"path":    req.URL.Path,
				"error":   errorMessage,
				"message": errorMessage,
			},
			"Config":  &AppConfig,
			"Route":   errorTemplatePath,
			"Message": errorMessage,
		})
		if err == nil {
			return
		}
		r.Logger.ErrorLog.Printf("Failed to execute error template %s: %v", errorTemplatePath, err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	errorHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Error %d</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; 
               margin: 0; padding: 2rem; line-height: 1.5; color: #333; max-width: 800px; margin: 0 auto; }
        h1 { font-size: 2rem; margin-bottom: 1rem; }
        .error-box { background-color: #f8f9fa; border-radius: 6px; padding: 2rem; 
                    box-shadow: 0 2px 15px rgba(0,0,0,0.05); margin: 2rem 0; border-left: 5px solid #dc3545; }
        pre { background: #f1f1f1; padding: 1rem; border-radius: 4px; overflow-x: auto; font-size: 0.9rem; }
        a { color: #007bff; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .back-link { margin-top: 2rem; display: inline-block; }
    </style>
</head>
<body>
    <h1>Error %d</h1>
    <div class="error-box">
        <p><strong>%s</strong></p>
        <p>Path: %s</p>
    </div>
    <a href="/" class="back-link">← Return to homepage</a>
</body>
</html>`, status, status, errorMessage, req.URL.Path)

	io.WriteString(w, errorHTML)
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	path = filepath.Clean(path)
	path = filepath.ToSlash(path)

	if path != "/" && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	if path == "" {
		return "/"
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

func matchPath(routePath, requestPath string) (map[string]string, bool) {
	routePath = normalizePath(routePath)
	requestPath = normalizePath(requestPath)
	params := make(map[string]string)

	if !strings.Contains(routePath, "[") {
		return params, routePath == requestPath
	}

	patternStr := "^" + paramRegex.ReplaceAllStringFunc(routePath, func(match string) string {
		return "([^/]+)"
	}) + "$"
	pattern := regexp.MustCompile(patternStr)

	matches := pattern.FindStringSubmatch(requestPath)
	if len(matches) == 0 {
		return nil, false
	}

	paramNames := paramRegex.FindAllStringSubmatch(routePath, -1)

	if len(matches)-1 != len(paramNames) {
		return nil, false
	}

	for i, paramMatch := range paramNames {
		if len(paramMatch) > 1 {
			params[paramMatch[1]] = matches[i+1]
		}
	}

	return params, true
}

func extractParamsFromRequest(requestPath, routePath string) map[string]string {
	params, _ := matchPath(routePath, requestPath)
	if params == nil {
		return make(map[string]string)
	}
	return params
}

func (r *Router) logRequest(req *http.Request, status int, duration time.Duration) {
	logLevel := AppConfig.LogLevel
	if logLevel == "debug" || logLevel == "info" {
		statusCode := status
		if statusCode == 0 {
			statusCode = 200
		}
		r.Logger.InfoLog.Printf("Handled: %s %s -> %d (%v)", req.Method, req.URL.Path, statusCode, duration.Round(time.Microsecond))
	} else if status >= 400 && logLevel != "error" {
		r.Logger.WarnLog.Printf("Handled: %s %s -> %d (%v)", req.Method, req.URL.Path, status, duration.Round(time.Microsecond))
	}
}
