package core

import (
	"html/template"
	"sync"
	"time"
)

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

	SSGCache      map[string]SSGCacheEntry
	SSGCacheDir   string
	ssgMutex      *sync.RWMutex
	ssgTaskChan   chan SSGTask
	ssgWorkerPool chan struct{}
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
		SSGCache:        make(map[string]SSGCacheEntry),
		SSGCacheDir:     ".goa/cache",
		ssgMutex:        &sync.RWMutex{},
	}
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

	m.ssgMutex.Lock()
	m.SSGCache = make(map[string]SSGCacheEntry)
	m.ssgMutex.Unlock()

	m.Logger.InfoLog.Printf("Template and SSG cache invalidated")
}
