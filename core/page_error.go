package core

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)


type PageError struct {
	Title    string
	Message  string
	Path     string
	Time     string
	Code     int
	Details  string
	Stack    string
	Original error
}


type APIError struct {
	Path     string
	Method   string
	ErrorMsg string
	Time     string
	Code     int
	Stack    string
	Original error
}


var pageErrorMap = sync.Map{}


var apiErrorMap = sync.Map{}


func RegisterPageError(routePath string, err error, code int) *PageError {
	if err == nil {
		return nil
	}

	details := err.Error()
	stack := string(debug.Stack())

	pe := &PageError{
		Title:    "Page Processing Error",
		Message:  fmt.Sprintf("The page '%s' could not be loaded due to an error", routePath),
		Path:     routePath,
		Time:     time.Now().Format(time.RFC3339),
		Code:     code,
		Details:  details,
		Stack:    stack,
		Original: err,
	}

	
	pageErrorMap.Store(routePath, pe)
	return pe
}


func RegisterAPIError(path string, method string, err error, code int) *APIError {
	if err == nil {
		return nil
	}

	apiErr := &APIError{
		Path:     path,
		Method:   method,
		ErrorMsg: err.Error(),
		Time:     time.Now().Format(time.RFC3339),
		Code:     code,
		Stack:    string(debug.Stack()),
		Original: err,
	}

	
	key := fmt.Sprintf("%s:%s", method, path)
	apiErrorMap.Store(key, apiErr)
	return apiErr
}


func GetPageError(routePath string) *PageError {
	if value, exists := pageErrorMap.Load(routePath); exists {
		if pe, ok := value.(*PageError); ok {
			return pe
		}
	}
	return nil
}


func GetAPIError(path string, method string) *APIError {
	key := fmt.Sprintf("%s:%s", method, path)
	if value, exists := apiErrorMap.Load(key); exists {
		if apiErr, ok := value.(*APIError); ok {
			return apiErr
		}
	}
	return nil
}


func ClearPageError(routePath string) {
	pageErrorMap.Delete(routePath)
}


func ClearAPIError(path string, method string) {
	key := fmt.Sprintf("%s:%s", method, path)
	apiErrorMap.Delete(key)
}


func HasPageError(routePath string) bool {
	_, exists := pageErrorMap.Load(routePath)
	return exists
}


func HasAPIError(path string, method string) bool {
	key := fmt.Sprintf("%s:%s", method, path)
	_, exists := apiErrorMap.Load(key)
	return exists
}


func (r *Router) RenderErrorPage(w http.ResponseWriter, req *http.Request, routePath string) {
	pageError := GetPageError(routePath)
	if pageError == nil {
		pageError = &PageError{
			Title:   "Unknown Error",
			Message: "An unknown error occurred while processing this page",
			Path:    routePath,
			Time:    time.Now().Format(time.RFC3339),
			Code:    http.StatusInternalServerError,
		}
	}

	
	errorTemplatePath := filepath.Join("core", "page", "error.html")
	errorTemplate, err := template.ParseFiles(errorTemplatePath)

	if err != nil {
		r.Logger.ErrorLog.Printf("Failed to parse error template: %v", err)
		r.renderFallbackErrorPage(w, pageError)
		return
	}

	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(pageError.Code)

	
	err = errorTemplate.Execute(w, map[string]interface{}{
		"Error":  pageError,
		"Config": &AppConfig,
		"Route":  routePath,
	})

	if err != nil {
		r.Logger.ErrorLog.Printf("Failed to execute error template: %v", err)
		r.renderFallbackErrorPage(w, pageError)
	}
}


func (r *Router) renderFallbackErrorPage(w http.ResponseWriter, pageError *PageError) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(pageError.Code)

	
	detailsSection := ""
	if pageError.Details != "" {
		detailsSection = fmt.Sprintf("<pre>%s</pre>", pageError.Details)
	}

	
	errorHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Error</title>
    <style>
        body { font-family: sans-serif; line-height: 1.6; color: #333; max-width: 800px; margin: 0 auto; padding: 1rem; }
        .error { background: #f8d7da; border-left: 5px solid #dc3545; padding: 1rem; margin: 1rem 0; }
        pre { background: #f5f5f5; padding: 1rem; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>Page Error</h1>
    <div class="error">
        <h2>%s</h2>
        <p>%s</p>
        <p><strong>Path:</strong> %s</p>
        <p><strong>Time:</strong> %s</p>
    </div>
    %s
    <p><a href="/">Return to Home</a></p>
</body>
</html>`,
		pageError.Title,
		pageError.Message,
		pageError.Path,
		pageError.Time,
		detailsSection,
	)

	io.WriteString(w, errorHTML)
}
