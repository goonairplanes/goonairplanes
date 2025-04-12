package core

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var jsLibrariesLoaded sync.Once
var jsLibraryMutex sync.RWMutex


func FetchAndCacheJSLibraries() error {
	if !AppConfig.InMemoryJS {
		return nil
	}

	var loadErr error
	jsLibrariesLoaded.Do(func() {
		var wg sync.WaitGroup

		
		libraries := map[string]string{
			"jquery": AppConfig.JQueryCDN,
			"alpine": AppConfig.AlpineJSCDN,
		}

		
		errCh := make(chan error, len(libraries))

		for lib, url := range libraries {
			wg.Add(1)
			go func(libName, libURL string) {
				defer wg.Done()

				
				client := &http.Client{
					Timeout: 10 * time.Second,
				}
				resp, err := client.Get(libURL)
				if err != nil {
					errCh <- fmt.Errorf("failed to fetch %s from %s: %w", libName, libURL, err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					errCh <- fmt.Errorf("failed to fetch %s from %s: status code %d", libName, libURL, resp.StatusCode)
					return
				}

				
				content, err := io.ReadAll(resp.Body)
				if err != nil {
					errCh <- fmt.Errorf("failed to read %s content: %w", libName, err)
					return
				}

				
				jsLibraryMutex.Lock()
				AppConfig.JSLibraryCache[libName] = string(content)
				jsLibraryMutex.Unlock()
			}(lib, url)
		}

		wg.Wait()
		close(errCh)

		
		for err := range errCh {
			if err != nil {
				loadErr = err
				break
			}
		}
	})

	return loadErr
}


func GetJSLibraryContent(library string) (string, bool, string) {
	if !AppConfig.InMemoryJS {
		switch library {
		case "alpine":
			return "", false, AppConfig.AlpineJSCDN
		case "jquery":
			return "", false, AppConfig.JQueryCDN
		default:
			return "", false, ""
		}
	}

	jsLibraryMutex.RLock()
	defer jsLibraryMutex.RUnlock()

	content, exists := AppConfig.JSLibraryCache[library]
	if exists {
		return content, true, ""
	}

	
	switch library {
	case "alpine":
		return "", false, AppConfig.AlpineJSCDN
	case "jquery":
		return "", false, AppConfig.JQueryCDN
	default:
		return "", false, ""
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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
