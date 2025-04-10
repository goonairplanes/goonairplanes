package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	router        *Router
	watcher       *fsnotify.Watcher
	debounceTimer *time.Timer
	dirs          []string
	logger        *AppLogger
}

func NewFileWatcher(router *Router, logger *AppLogger) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &FileWatcher{
		router:  router,
		watcher: watcher,
		dirs:    []string{AppConfig.AppDir, AppConfig.StaticDir},
		logger:  logger,
	}, nil
}

func (fw *FileWatcher) Start() {
	fw.logger.InfoLog.Printf("Starting file watcher...")

	for _, dir := range fw.dirs {
		err := fw.watchDir(dir)
		if err != nil {
			fw.logger.ErrorLog.Printf("Error watching directory %s: %v", dir, err)
		} else {
			fw.logger.InfoLog.Printf("Watching directory: %s", dir)
		}
	}

	go fw.watchLoop()
}

func (fw *FileWatcher) Stop() {
	if fw.watcher != nil {
		fw.watcher.Close()
		fw.logger.InfoLog.Printf("File watcher stopped")
	}
}

func (fw *FileWatcher) watchLoop() {
	debounceTimeout := 100 * time.Millisecond

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				if fw.debounceTimer != nil {
					fw.debounceTimer.Stop()
				}
				fw.debounceTimer = time.AfterFunc(debounceTimeout, func() {
					fw.logger.InfoLog.Printf("File change detected in %s, reloading...", event.Name)
					err := fw.router.InitRoutes()
					if err != nil {
						fw.logger.ErrorLog.Printf("Failed to reload templates: %v", err)
					} else {
						fw.logger.InfoLog.Printf("Templates reloaded successfully")
					}
				})
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fw.logger.ErrorLog.Printf("Watcher error: %v", err)
		}
	}
}

func (fw *FileWatcher) watchDir(dir string) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return fw.watcher.Add(path)
		}
		return nil
	})
	return err
}
