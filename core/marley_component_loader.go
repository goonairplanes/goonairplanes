package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)


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
