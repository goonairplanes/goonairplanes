package core

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var metaTagRegex = regexp.MustCompile(`<!--meta:([a-zA-Z0-9_:,\-\s]+)-->`)
var renderModeRegex = regexp.MustCompile(`<!--render:([a-zA-Z]+)-->`)
var titleRegex = regexp.MustCompile(`<!--title:([^-]+)-->`)
var descRegex = regexp.MustCompile(`<!--description:([^-]+)-->`)


var htmlCommentMetaTagRegex = regexp.MustCompile(`<!---meta:([a-zA-Z0-9_:,\-\s]+)(?:-->|--->)`)
var htmlCommentRenderModeRegex = regexp.MustCompile(`<!---render:([a-zA-Z]+)(?:-->|--->)`)
var htmlCommentTitleRegex = regexp.MustCompile(`<!---title:([^-]+)(?:-->|--->)`)
var htmlCommentDescRegex = regexp.MustCompile(`<!---description:([^-]+)(?:-->|--->)`)

type PageMetadata struct {
	Title       string
	Description string
	MetaTags    map[string]string
	RenderMode  string
}

type Marley struct {
	Templates       map[string]*template.Template
	Components      map[string]*template.Template
	LayoutTemplate  *template.Template
	ComponentsCache map[string]string
	PageMetadata    map[string]*PageMetadata
	LayoutMetadata  *PageMetadata
	mutex           sync.RWMutex
	cacheExpiry     time.Time
	cacheTTL        time.Duration
	Logger          *AppLogger
	BundledAssets   map[string]string
	BundleMode      bool
	
	SSGCache    map[string]string
	SSGCacheDir string
}

func NewMarley(logger *AppLogger) *Marley {
	return &Marley{
		Templates:       make(map[string]*template.Template),
		Components:      make(map[string]*template.Template),
		ComponentsCache: make(map[string]string),
		PageMetadata:    make(map[string]*PageMetadata),
		LayoutMetadata:  nil,
		cacheTTL:        5 * time.Minute,
		Logger:          logger,
		BundledAssets:   make(map[string]string),
		BundleMode:      false,
		SSGCache:        make(map[string]string),
		SSGCacheDir:     ".goa/cache",
	}
}

func (m *Marley) LoadTemplates() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	if !m.cacheExpiry.IsZero() && now.Before(m.cacheExpiry) && AppConfig.TemplateCache {
		return nil
	}

	startTime := time.Now()
	m.Logger.InfoLog.Printf("Loading templates...")

	var wg sync.WaitGroup
	errorCh := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := m.loadComponents(); err != nil {
			errorCh <- err
		}
	}()

	if m.BundleMode {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := m.bundleAssets(); err != nil {
				errorCh <- err
			}
		}()
	}

	layoutCh := make(chan []byte)
	layoutErrCh := make(chan error, 1)

	go func() {
		layoutContent, err := os.ReadFile(AppConfig.LayoutPath)
		if err != nil {
			layoutErrCh <- fmt.Errorf("failed to load layout template: %w", err)
			return
		}
		layoutCh <- layoutContent
	}()

	wg.Wait()

	select {
	case err := <-errorCh:
		m.Logger.ErrorLog.Printf("Failed to load components: %v", err)
		return err
	default:
	}

	var layoutContent []byte
	select {
	case err := <-layoutErrCh:
		m.Logger.ErrorLog.Printf("Failed to load layout template: %v", err)
		return err
	case layoutContent = <-layoutCh:
		m.Logger.InfoLog.Printf("Layout template loaded successfully")

		
		m.LayoutMetadata = extractPageMetadata(string(layoutContent), "layout")
		m.Logger.InfoLog.Printf("Layout metadata extracted: %s", m.LayoutMetadata.Title)
	}

	var (
		templatePaths []string
		mu            sync.Mutex
	)

	err := filepath.Walk(AppConfig.AppDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".html" &&
			path != AppConfig.LayoutPath &&
			!strings.HasPrefix(path, AppConfig.ComponentDir) {

			routePath := getRoutePathFromFile(path, AppConfig.AppDir)

			if routePath == "layout" {
				return nil
			}

			if strings.HasPrefix(routePath, "components/") {
				return nil
			}

			mu.Lock()
			templatePaths = append(templatePaths, path)
			mu.Unlock()
		}

		return nil
	})
	if err != nil {
		m.Logger.ErrorLog.Printf("Failed to scan template directories: %v", err)
		return err
	}

	templates := make(map[string]*template.Template)
	pageMetadata := make(map[string]*PageMetadata)

	templateCh := make(chan struct {
		path     string
		tmpl     *template.Template
		metadata *PageMetadata
	}, len(templatePaths))

	semaphore := make(chan struct{}, 4)
	errCh := make(chan error, len(templatePaths))

	var collectorWg sync.WaitGroup
	collectorWg.Add(1)

	go func() {
		defer collectorWg.Done()
		for i := 0; i < len(templatePaths); i++ {
			result := <-templateCh
			templates[result.path] = result.tmpl
			pageMetadata[result.path] = result.metadata

			if AppConfig.SSGEnabled && result.metadata.RenderMode == "ssg" {
				if err := m.generateStaticFile(result.path, result.tmpl, result.metadata); err != nil {
					m.Logger.WarnLog.Printf("Failed to generate static file for %s: %v", result.path, err)
				} else {
					m.Logger.InfoLog.Printf("Generated static file for %s", result.path)
				}
			}
		}
	}()

	for _, path := range templatePaths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			routePath := getRoutePathFromFile(p, AppConfig.AppDir)

			pageContent, err := os.ReadFile(p)
			if err != nil {
				errCh <- fmt.Errorf("failed to read template %s: %w", p, err)
				return
			}

			metadata := extractPageMetadata(string(pageContent), routePath)

			processedContent := processPageContent(string(pageContent), metadata)

			tmpl := template.New("layout")

			_, err = tmpl.Parse(string(layoutContent))
			if err != nil {
				errCh <- fmt.Errorf("failed to parse layout template: %w", err)
				return
			}

			for name, content := range m.ComponentsCache {
				_, err = tmpl.New(name).Parse(content)
				if err != nil {
					errCh <- fmt.Errorf("failed to parse component %s: %w", name, err)
					return
				}
			}

			_, err = tmpl.New("page").Parse(processedContent)
			if err != nil {
				errCh <- fmt.Errorf("failed to parse template %s: %w", p, err)
				return
			}

			templateCh <- struct {
				path     string
				tmpl     *template.Template
				metadata *PageMetadata
			}{routePath, tmpl, metadata}

			m.Logger.InfoLog.Printf("Template loaded: %s → %s (mode: %s)", p, routePath, metadata.RenderMode)
		}(path)
	}

	wg.Wait()
	close(errCh)

	close(templateCh)

	collectorWg.Wait()

	for err := range errCh {
		if err != nil {
			m.Logger.ErrorLog.Printf("Template processing error: %v", err)
			return err
		}
	}

	m.Templates = templates
	m.PageMetadata = pageMetadata

	if AppConfig.TemplateCache {
		m.cacheExpiry = now.Add(m.cacheTTL)
	}

	elapsedTime := time.Since(startTime)
	m.Logger.InfoLog.Printf("Templates loaded successfully in %v", elapsedTime.Round(time.Millisecond))

	return nil
}

func extractPageMetadata(content, _ string) *PageMetadata {
	
	metadata := &PageMetadata{
		Title:       "Go on Airplanes",
		Description: AppConfig.DefaultMetaTags["description"],
		MetaTags:    make(map[string]string),
		RenderMode:  AppConfig.DefaultRenderMode,
	}

	
	for k, v := range AppConfig.DefaultMetaTags {
		metadata.MetaTags[k] = v
	}

	
	foundTitle := false
	foundDesc := false
	foundRenderMode := false

	
	htmlTitleMatch := htmlCommentTitleRegex.FindStringSubmatch(content)
	if len(htmlTitleMatch) > 1 {
		titleText := strings.TrimSpace(htmlTitleMatch[1])
		
		titleText = strings.TrimSuffix(titleText, "-")
		titleText = strings.TrimSpace(titleText)
		metadata.Title = titleText
		metadata.MetaTags["og:title"] = titleText
		foundTitle = true
	}

	htmlDescMatch := htmlCommentDescRegex.FindStringSubmatch(content)
	if len(htmlDescMatch) > 1 {
		descText := strings.TrimSpace(htmlDescMatch[1])
		
		descText = strings.TrimSuffix(descText, "-")
		descText = strings.TrimSpace(descText)
		metadata.Description = descText
		metadata.MetaTags["description"] = descText
		metadata.MetaTags["og:description"] = descText
		foundDesc = true
	}

	htmlMetaMatches := htmlCommentMetaTagRegex.FindAllStringSubmatch(content, -1)
	for _, match := range htmlMetaMatches {
		if len(match) > 1 {
			metaText := strings.TrimSpace(match[1])
			
			metaText = strings.TrimSuffix(metaText, "-")
			metaText = strings.TrimSpace(metaText)
			parts := strings.SplitN(metaText, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				metadata.MetaTags[key] = value
			}
		}
	}

	htmlRenderMatch := htmlCommentRenderModeRegex.FindStringSubmatch(content)
	if len(htmlRenderMatch) > 1 {
		renderText := strings.TrimSpace(htmlRenderMatch[1])
		
		renderText = strings.TrimSuffix(renderText, "-")
		renderText = strings.TrimSpace(renderText)
		mode := strings.ToLower(renderText)
		if mode == "ssr" || mode == "ssg" {
			metadata.RenderMode = mode
			foundRenderMode = true
		}
	}

	
	if !foundTitle {
		titleMatch := titleRegex.FindStringSubmatch(content)
		if len(titleMatch) > 1 {
			metadata.Title = strings.TrimSpace(titleMatch[1])
			metadata.MetaTags["og:title"] = metadata.Title
		}
	}

	if !foundDesc {
		descMatch := descRegex.FindStringSubmatch(content)
		if len(descMatch) > 1 {
			metadata.Description = strings.TrimSpace(descMatch[1])
			metadata.MetaTags["description"] = metadata.Description
			metadata.MetaTags["og:description"] = metadata.Description
		}
	}

	
	metaMatches := metaTagRegex.FindAllStringSubmatch(content, -1)
	for _, match := range metaMatches {
		if len(match) > 1 {
			parts := strings.SplitN(match[1], ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				if _, exists := metadata.MetaTags[key]; !exists {
					metadata.MetaTags[key] = value
				}
			}
		}
	}

	if !foundRenderMode {
		renderMatch := renderModeRegex.FindStringSubmatch(content)
		if len(renderMatch) > 1 {
			mode := strings.ToLower(renderMatch[1])
			if mode == "ssr" || mode == "ssg" {
				metadata.RenderMode = mode
			}
		}
	}

	return metadata
}

func processPageContent(content string, _ *PageMetadata) string {
	
	contentBefore := len(content)

	
	
	content = titleRegex.ReplaceAllString(content, "")
	content = descRegex.ReplaceAllString(content, "")
	content = metaTagRegex.ReplaceAllString(content, "")
	content = renderModeRegex.ReplaceAllString(content, "")

	
	content = htmlCommentTitleRegex.ReplaceAllString(content, "")
	content = htmlCommentDescRegex.ReplaceAllString(content, "")
	content = htmlCommentMetaTagRegex.ReplaceAllString(content, "")
	content = htmlCommentRenderModeRegex.ReplaceAllString(content, "")

	
	content = strings.TrimLeft(content, "\r\n")

	contentAfter := len(content)
	bytesRemoved := contentBefore - contentAfter

	if bytesRemoved > 0 {
		
	}

	return content
}



func (m *Marley) generateStaticFile(routePath string, tmpl *template.Template, metadata *PageMetadata) error {
	if !AppConfig.SSGEnabled {
		return nil
	}

	
	finalMetadata := m.mergeMetadata(routePath, metadata)

	relativePath := strings.TrimPrefix(routePath, "/")
	if routePath == "/" {
		relativePath = "index"
	}

	
	var buffer strings.Builder

	
	now := time.Now()
	templateData := map[string]interface{}{
		"Metadata":    finalMetadata,
		"Config":      &AppConfig,
		"BuildTime":   now.Format(time.RFC1123),
		"ServerTime":  now.Format(time.RFC1123),
		"CurrentTime": now,
		"Route":       routePath,
	}

	
	if m.BundleMode {
		templateData["Bundles"] = m.BundledAssets
	}

	
	err := tmpl.ExecuteTemplate(&buffer, "layout", templateData)
	if err != nil {
		return fmt.Errorf("failed to render template to memory: %w", err)
	}

	
	content := buffer.String()

	
	
	
	
	m.SSGCache[routePath] = content

	
	if AppConfig.SSGCacheEnabled {
		cacheDir := m.SSGCacheDir
		fullPath := filepath.Join(cacheDir, relativePath+".html")

		dirPath := filepath.Dir(fullPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			m.Logger.WarnLog.Printf("Failed to create cache directory: %v", err)
			return nil 
		}

		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			m.Logger.WarnLog.Printf("Failed to write to cache file: %v", err)
			
		} else {
			m.Logger.InfoLog.Printf("Cached SSG content to disk: %s", fullPath)
		}
	}

	m.Logger.InfoLog.Printf("Generated in-memory SSG content for: %s (mode: %s, size: %d bytes)",
		routePath, finalMetadata.RenderMode, buffer.Len())
	return nil
}

func (m *Marley) loadComponents() error {
	componentCache := make(map[string]string)
	componentDir := AppConfig.ComponentDir

	err := filepath.Walk(componentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".html" {
			componentContent, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read component %s: %w", path, err)
			}

			componentName := strings.TrimSuffix(filepath.Base(path), ".html")
			componentCache[componentName] = string(componentContent)
		}

		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load components: %w", err)
	}

	m.ComponentsCache = componentCache
	return nil
}

func getRoutePathFromFile(fullPath, basePath string) string {
	fullPath = filepath.ToSlash(fullPath)
	basePath = filepath.ToSlash(basePath)

	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	relativePath := fullPath
	if strings.HasPrefix(fullPath, basePath) {
		relativePath = strings.TrimPrefix(fullPath, basePath)
	}

	relativePath = strings.TrimSuffix(relativePath, ".html")

	if relativePath == "index" {
		return "/"
	} else if strings.HasSuffix(relativePath, "/index") {
		relativePath = strings.TrimSuffix(relativePath, "/index")
		if relativePath == "" {
			return "/"
		}
	}

	if relativePath != "/" && !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}

	return relativePath
}

func (m *Marley) bundleAssets() error {
	assetTypes := map[string]string{
		".css": "css",
		".js":  "js",
	}

	for ext, assetType := range assetTypes {
		bundleName := fmt.Sprintf("bundle.%s", assetType)
		bundlePath := filepath.Join(AppConfig.StaticDir, bundleName)

		var bundleContent strings.Builder
		assetFiles := make([]string, 0)

		err := filepath.Walk(AppConfig.StaticDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && filepath.Ext(path) == ext && !strings.Contains(path, "bundle.") {
				assetFiles = append(assetFiles, path)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to scan static directory for %s files: %w", assetType, err)
		}

		if len(assetFiles) == 0 {
			continue
		}

		for _, assetPath := range assetFiles {
			content, err := os.ReadFile(assetPath)
			if err != nil {
				return fmt.Errorf("failed to read asset file %s: %w", assetPath, err)
			}

			relPath, _ := filepath.Rel(AppConfig.StaticDir, assetPath)
			bundleContent.WriteString(fmt.Sprintf("/* %s */\n", relPath))
			bundleContent.Write(content)
			bundleContent.WriteString("\n\n")
		}

		err = os.WriteFile(bundlePath, []byte(bundleContent.String()), 0644)
		if err != nil {
			return fmt.Errorf("failed to write bundle file %s: %w", bundlePath, err)
		}

		m.BundledAssets[assetType] = fmt.Sprintf("/static/%s", bundleName)
		m.Logger.InfoLog.Printf("Created %s bundle with %d files", assetType, len(assetFiles))
	}

	return nil
}

func (m *Marley) mergeMetadata(routePath string, pageMetadata *PageMetadata) *PageMetadata {
	
	result := &PageMetadata{
		Title:       "Go on Airplanes",                        
		Description: AppConfig.DefaultMetaTags["description"], 
		MetaTags:    make(map[string]string),
		RenderMode:  AppConfig.DefaultRenderMode, 
	}

	
	for k, v := range AppConfig.DefaultMetaTags {
		
		if k != "description" && k != "og:description" && k != "og:title" {
			result.MetaTags[k] = v
		}
	}

	
	if m.LayoutMetadata != nil {
		
		if m.LayoutMetadata.Title != "Go on Airplanes" {
			result.Title = m.LayoutMetadata.Title
		}

		
		if m.LayoutMetadata.Description != AppConfig.DefaultMetaTags["description"] {
			result.Description = m.LayoutMetadata.Description
		}

		
		for k, v := range m.LayoutMetadata.MetaTags {
			if k != "description" && k != "og:description" && k != "og:title" {
				result.MetaTags[k] = v
			}
		}

		
		if m.LayoutMetadata.RenderMode != AppConfig.DefaultRenderMode {
			result.RenderMode = m.LayoutMetadata.RenderMode
		}
	}

	
	
	if pageMetadata.Title != "Go on Airplanes" {
		result.Title = pageMetadata.Title
	}

	
	if pageMetadata.Description != AppConfig.DefaultMetaTags["description"] {
		result.Description = pageMetadata.Description
	}

	
	for k, v := range pageMetadata.MetaTags {
		if k != "description" && k != "og:description" && k != "og:title" {
			result.MetaTags[k] = v
		}
	}

	
	if pageMetadata.RenderMode != AppConfig.DefaultRenderMode {
		result.RenderMode = pageMetadata.RenderMode
	}

	
	
	result.MetaTags["og:title"] = result.Title

	
	if result.Description != "" {
		result.MetaTags["description"] = result.Description
		result.MetaTags["og:description"] = result.Description
	}

	m.Logger.InfoLog.Printf("Merged metadata for %s: Title='%s', Description='%s...', Mode='%s'",
		routePath,
		result.Title,
		truncateString(result.Description, 30),
		result.RenderMode)

	return result
}


func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}



func (m *Marley) RenderTemplate(w http.ResponseWriter, route string, data interface{}) error {
	m.mutex.RLock()
	tmpl, exists := m.Templates[route]
	metadata, metadataExists := m.PageMetadata[route]

	
	cachedContent, hasCachedContent := m.SSGCache[route]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("template for route %s not found", route)
	}

	var finalMetadata *PageMetadata
	if metadataExists {
		
		finalMetadata = m.mergeMetadata(route, metadata)
	} else {
		
		if m.LayoutMetadata != nil {
			finalMetadata = m.LayoutMetadata
		} else {
			
			finalMetadata = &PageMetadata{
				Title:       "Go on Airplanes",
				Description: AppConfig.DefaultMetaTags["description"],
				MetaTags:    AppConfig.DefaultMetaTags,
				RenderMode:  AppConfig.DefaultRenderMode,
			}
		}
	}

	
	if AppConfig.SSGEnabled && finalMetadata.RenderMode == "ssg" {
		if hasCachedContent {
			
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(cachedContent))
			m.Logger.InfoLog.Printf("Served SSG content from memory cache: %s", route)
			return nil
		} else {
			
			
			if err := m.generateStaticFile(route, tmpl, metadata); err != nil {
				m.Logger.WarnLog.Printf("Failed to generate SSG content for %s: %v", route, err)
				
			} else {
				
				m.mutex.RLock()
				cachedContent, hasCachedContent = m.SSGCache[route]
				m.mutex.RUnlock()

				if hasCachedContent {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.Write([]byte(cachedContent))
					m.Logger.InfoLog.Printf("Served freshly generated SSG content: %s", route)
					return nil
				}
			}
		}
	}

	
	now := time.Now()
	templateData := map[string]interface{}{
		"Metadata":    finalMetadata,
		"Config":      &AppConfig,
		"ServerTime":  now.Format(time.RFC1123),
		"CurrentTime": now,
		"Route":       route,
	}

	
	if m.BundleMode {
		templateData["Bundles"] = m.BundledAssets
	}

	
	if existingData, ok := data.(map[string]interface{}); ok {
		for k, v := range existingData {
			templateData[k] = v
		}
	} else if data != nil {
		templateData["Data"] = data
	}

	
	m.Logger.InfoLog.Printf("Rendering template %s with mode: %s, title: %s",
		route, finalMetadata.RenderMode, finalMetadata.Title)

	return tmpl.ExecuteTemplate(w, "layout", templateData)
}

func (m *Marley) SetCacheTTL(duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.cacheTTL = duration
	m.cacheExpiry = time.Time{}
	m.Logger.InfoLog.Printf("Template cache TTL set to %v", duration)
}

func (m *Marley) InvalidateCache() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.cacheExpiry = time.Time{}
	
	m.SSGCache = make(map[string]string)
	m.Logger.InfoLog.Printf("Template and SSG cache invalidated")
}
