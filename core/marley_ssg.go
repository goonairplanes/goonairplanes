package core

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)


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
