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

var metaTagRegex = regexp.MustCompile(`<!--meta:([a-zA-Z0-9_:]+)-->`)
var renderModeRegex = regexp.MustCompile(`<!--render:([a-zA-Z]+)-->`)
var titleRegex = regexp.MustCompile(`<!--title:([^-]+)-->`)
var descRegex = regexp.MustCompile(`<!--description:([^-]+)-->`)

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
	mutex           sync.RWMutex
	cacheExpiry     time.Time
	cacheTTL        time.Duration
	Logger          *AppLogger
	BundledAssets   map[string]string
	BundleMode      bool
}

func NewMarley(logger *AppLogger) *Marley {
	return &Marley{
		Templates:       make(map[string]*template.Template),
		Components:      make(map[string]*template.Template),
		ComponentsCache: make(map[string]string),
		PageMetadata:    make(map[string]*PageMetadata),
		cacheTTL:        5 * time.Minute,
		Logger:          logger,
		BundledAssets:   make(map[string]string),
		BundleMode:      false,
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

	titleMatch := titleRegex.FindStringSubmatch(content)
	if len(titleMatch) > 1 {
		metadata.Title = strings.TrimSpace(titleMatch[1])
		metadata.MetaTags["og:title"] = metadata.Title
	}

	descMatch := descRegex.FindStringSubmatch(content)
	if len(descMatch) > 1 {
		metadata.Description = strings.TrimSpace(descMatch[1])
		metadata.MetaTags["description"] = metadata.Description
		metadata.MetaTags["og:description"] = metadata.Description
	}

	metaMatches := metaTagRegex.FindAllStringSubmatch(content, -1)
	for _, match := range metaMatches {
		if len(match) > 1 {
			parts := strings.SplitN(match[1], ":", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				metadata.MetaTags[key] = value
			}
		}
	}

	renderMatch := renderModeRegex.FindStringSubmatch(content)
	if len(renderMatch) > 1 {
		mode := strings.ToLower(renderMatch[1])
		if mode == "ssr" || mode == "ssg" {
			metadata.RenderMode = mode
		}
	}

	return metadata
}

func processPageContent(content string, _ *PageMetadata) string {

	content = titleRegex.ReplaceAllString(content, "")
	content = descRegex.ReplaceAllString(content, "")
	content = metaTagRegex.ReplaceAllString(content, "")
	content = renderModeRegex.ReplaceAllString(content, "")

	return content
}

func (m *Marley) generateStaticFile(routePath string, tmpl *template.Template, metadata *PageMetadata) error {
	if !AppConfig.SSGEnabled {
		return nil
	}

	relativePath := strings.TrimPrefix(routePath, "/")
	if routePath == "/" {
		relativePath = "index"
	}

	staticDir := AppConfig.SSGDir
	fullPath := filepath.Join(staticDir, relativePath+".html")

	dirPath := filepath.Dir(fullPath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory for static file: %w", err)
	}

	m.Logger.InfoLog.Printf("Generating static file at: %s", fullPath)

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create static file: %w", err)
	}
	defer file.Close()

	data := map[string]interface{}{
		"Metadata":  metadata,
		"Config":    &AppConfig,
		"BuildTime": time.Now().Format(time.RFC1123),
	}

	err = tmpl.ExecuteTemplate(file, "layout", data)
	if err != nil {
		return fmt.Errorf("failed to render template to static file: %w", err)
	}

	m.Logger.InfoLog.Printf("Successfully generated static file: %s", fullPath)
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

func (m *Marley) RenderTemplate(w http.ResponseWriter, route string, data interface{}) error {
	m.mutex.RLock()
	tmpl, exists := m.Templates[route]
	metadata, metadataExists := m.PageMetadata[route]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("template for route %s not found", route)
	}

	if AppConfig.SSGEnabled && metadataExists && metadata.RenderMode == "ssg" {
		staticPath := route
		if staticPath == "/" {
			staticPath = "/index"
		}

		ssgFilePath := filepath.Join(AppConfig.SSGDir, strings.TrimPrefix(staticPath, "/")) + ".html"
		if _, err := os.Stat(ssgFilePath); err == nil {
			w.Header().Set("Location", "/static/generated"+staticPath+".html")
			w.WriteHeader(http.StatusTemporaryRedirect)
			return nil
		} else {
			m.Logger.WarnLog.Printf("Static file for route %s not found, falling back to SSR", route)
		}
	}

	if metadataExists {
		var combinedData map[string]interface{}

		if existingData, ok := data.(map[string]interface{}); ok {
			combinedData = existingData
			combinedData["Metadata"] = metadata
			combinedData["Config"] = &AppConfig
		} else {
			combinedData = map[string]interface{}{
				"Data":     data,
				"Metadata": metadata,
				"Config":   &AppConfig,
			}
		}

		if m.BundleMode {
			combinedData["Bundles"] = m.BundledAssets
		}

		return tmpl.ExecuteTemplate(w, "layout", combinedData)
	}

	return tmpl.ExecuteTemplate(w, "layout", data)
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
	m.Logger.InfoLog.Printf("Template cache invalidated")
}
