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

	templateErrors := make(map[string]error)
	var templateErrorsMutex sync.Mutex

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
		err      error
	}, len(templatePaths))

	semaphore := make(chan struct{}, 4)

	var collectorWg sync.WaitGroup
	collectorWg.Add(1)

	go func() {
		defer collectorWg.Done()
		for i := 0; i < len(templatePaths); i++ {
			result := <-templateCh

			if result.err != nil {
				templateErrorsMutex.Lock()
				templateErrors[result.path] = result.err
				templateErrorsMutex.Unlock()

				RegisterPageError(result.path, result.err, http.StatusInternalServerError)

				m.Logger.ErrorLog.Printf("Failed to load template %s: %v", result.path, result.err)
				continue
			}

			templates[result.path] = result.tmpl
			pageMetadata[result.path] = result.metadata

			ClearPageError(result.path)

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

			defer func() {
				if r := recover(); r != nil {
					err := fmt.Errorf("panic during template processing: %v", r)
					templateCh <- struct {
						path     string
						tmpl     *template.Template
						metadata *PageMetadata
						err      error
					}{routePath, nil, nil, err}
				}
			}()

			pageContent, err := os.ReadFile(p)
			if err != nil {
				templateCh <- struct {
					path     string
					tmpl     *template.Template
					metadata *PageMetadata
					err      error
				}{routePath, nil, nil, fmt.Errorf("failed to read template %s: %w", p, err)}
				return
			}

			metadata := extractPageMetadata(string(pageContent), routePath)

			processedContent := processPageContent(string(pageContent), metadata)

			tmpl := template.New("layout")

			_, err = tmpl.Parse(string(layoutContent))
			if err != nil {
				templateCh <- struct {
					path     string
					tmpl     *template.Template
					metadata *PageMetadata
					err      error
				}{routePath, nil, nil, fmt.Errorf("failed to parse layout template: %w", err)}
				return
			}

			for name, content := range m.ComponentsCache {
				_, err = tmpl.New(name).Parse(content)
				if err != nil {
					templateCh <- struct {
						path     string
						tmpl     *template.Template
						metadata *PageMetadata
						err      error
					}{routePath, nil, nil, fmt.Errorf("failed to parse component %s for template %s: %w", name, p, err)}
					return
				}
			}

			_, err = tmpl.New("page").Parse(processedContent)
			if err != nil {
				templateCh <- struct {
					path     string
					tmpl     *template.Template
					metadata *PageMetadata
					err      error
				}{routePath, nil, nil, fmt.Errorf("failed to parse template %s: %w", p, err)}
				return
			}

			templateCh <- struct {
				path     string
				tmpl     *template.Template
				metadata *PageMetadata
				err      error
			}{routePath, tmpl, metadata, nil}

			m.Logger.InfoLog.Printf("Template loaded: %s → %s (mode: %s)", p, routePath, metadata.RenderMode)
		}(path)
	}

	wg.Wait()
	close(templateCh)

	collectorWg.Wait()

	m.TemplateErrors = templateErrors

	if len(templateErrors) > 0 {
		m.Logger.WarnLog.Printf("%d templates failed to load", len(templateErrors))
		for path, err := range templateErrors {
			m.Logger.WarnLog.Printf("  - %s: %v", path, err)
		}
	}

	m.Templates = templates
	m.PageMetadata = pageMetadata

	if AppConfig.TemplateCache {
		m.cacheExpiry = now.Add(m.cacheTTL)
	}

	elapsedTime := time.Since(startTime)
	m.Logger.InfoLog.Printf("Templates loaded successfully in %v (%d templates, %d failed)",
		elapsedTime.Round(time.Millisecond), len(templates), len(templateErrors))

	return nil
}
