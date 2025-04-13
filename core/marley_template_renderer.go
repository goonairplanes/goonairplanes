package core

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var renderCache = &sync.Map{}

func (m *Marley) RenderTemplate(w http.ResponseWriter, route string, data interface{}) error {
	startTime := time.Now()
	m.mutex.RLock()
	tmpl, ok := m.Templates[route]
	metadata, metaOk := m.PageMetadata[route]

	templateErr, hasError := m.TemplateErrors[route]
	m.mutex.RUnlock()

	if hasError {
		m.Logger.ErrorLog.Printf("Attempted to render template with known errors: %s: %v", route, templateErr)

		if !HasPageError(route) {
			RegisterPageError(route, templateErr, http.StatusInternalServerError)
		}
		return fmt.Errorf("template has loading errors: %w", templateErr)
	}

	if !ok {
		err := fmt.Errorf("template not found: %s", route)
		RegisterPageError(route, err, http.StatusNotFound)
		return err
	}

	if !metaOk {
		metadata = &PageMetadata{
			Title:       defaultTitle,
			Description: AppConfig.DefaultMetaTags["description"],
			MetaTags:    make(map[string]string),
			RenderMode:  AppConfig.DefaultRenderMode,
			JSLibrary:   defaultJSLibrary,
		}
	}

	finalMetadata := m.mergeMetadata(route, metadata)

	if cachedContent := m.GetCachedSSGContent(route); cachedContent != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-SSG-Cached", "true")
		io.WriteString(w, cachedContent)
		return nil
	}

	cacheKey := "rendered:" + route
	if cachedHTML, found := renderCache.Load(cacheKey); found {
		if renderedHTML, ok := cachedHTML.(string); ok {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("X-Template-Cached", "true")
			io.WriteString(w, renderedHTML)
			return nil
		}
	}

	if AppConfig.LogLevel != "error" {
		m.Logger.InfoLog.Printf("Rendering template %s with mode: %s, title: %s, js: %s",
			route, finalMetadata.RenderMode, finalMetadata.Title, finalMetadata.JSLibrary)
	}

	var buffer strings.Builder
	buffer.Grow(16 * 1024)

	now := time.Now()
	templateData := map[string]interface{}{
		"Metadata":    finalMetadata,
		"Config":      &AppConfig,
		"BuildTime":   now.Format(time.RFC1123),
		"ServerTime":  now.Format(time.RFC1123),
		"CurrentTime": now,
		"Route":       route,
		"Data":        data,
	}

	if m.BundleMode {
		templateData["Bundles"] = m.BundledAssets
	}

	defer func() {
		if r := recover(); r != nil {
			m.Logger.ErrorLog.Printf("Panic during template execution: %v", r)
			err := fmt.Errorf("template execution panic: %v", r)

			m.mutex.Lock()
			m.TemplateErrors[route] = err
			m.mutex.Unlock()

			RegisterPageError(route, err, http.StatusInternalServerError)
		}
	}()

	err := tmpl.ExecuteTemplate(&buffer, "layout", templateData)
	if err != nil {
		m.Logger.ErrorLog.Printf("Error executing template %s: %v", route, err)

		m.mutex.Lock()
		m.TemplateErrors[route] = err
		m.mutex.Unlock()

		RegisterPageError(route, err, http.StatusInternalServerError)

		return fmt.Errorf("error rendering template: %w", err)
	}

	renderedHTML := buffer.String()

	renderedHTML = injectJavaScriptLibraries(renderedHTML, finalMetadata.JSLibrary)

	if AppConfig.TemplateCache && len(renderedHTML) < 64*1024 {
		renderCache.Store(cacheKey, renderedHTML)
	}

	if finalMetadata.RenderMode == "ssg" && AppConfig.SSGEnabled {
		go func() {
			if err := m.generateStaticFile(route, tmpl, metadata); err != nil {
				m.Logger.WarnLog.Printf("Failed to generate static file for %s: %v", route, err)
			}
		}()
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, renderedHTML)

	renderTime := time.Since(startTime)
	if renderTime > 5*time.Millisecond && AppConfig.LogLevel == "debug" {
		m.Logger.InfoLog.Printf("Slow template render: %s took %v", route, renderTime)
	}

	return nil
}
