package core

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Marley struct {
	Templates       map[string]*template.Template
	Components      map[string]*template.Template
	LayoutTemplate  *template.Template
	ComponentsCache map[string]string
	mutex           sync.RWMutex
	cacheExpiry     time.Time
	cacheTTL        time.Duration
	Logger          *AppLogger
}

func NewMarley(logger *AppLogger) *Marley {
	return &Marley{
		Templates:       make(map[string]*template.Template),
		Components:      make(map[string]*template.Template),
		ComponentsCache: make(map[string]string),
		cacheTTL:        5 * time.Minute,
		Logger:          logger,
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
	semaphore := make(chan struct{}, 4)
	errCh := make(chan error, len(templatePaths))

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

			_, err = tmpl.New("page").Parse(string(pageContent))
			if err != nil {
				errCh <- fmt.Errorf("failed to parse template %s: %w", p, err)
				return
			}

			mu.Lock()
			templates[routePath] = tmpl
			mu.Unlock()

			m.Logger.InfoLog.Printf("Template loaded: %s → %s", p, routePath)
		}(path)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			m.Logger.ErrorLog.Printf("Template processing error: %v", err)
			return err
		}
	}

	m.Templates = templates

	if AppConfig.TemplateCache {
		m.cacheExpiry = now.Add(m.cacheTTL)
	}

	elapsedTime := time.Since(startTime)
	m.Logger.InfoLog.Printf("Templates loaded successfully in %v", elapsedTime.Round(time.Millisecond))

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

func (m *Marley) RenderTemplate(w http.ResponseWriter, route string, data interface{}) error {
	m.mutex.RLock()
	tmpl, exists := m.Templates[route]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("template for route %s not found", route)
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
