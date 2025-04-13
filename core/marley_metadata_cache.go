package core

import (
	"sync"
	"time"
)


type MetadataCache struct {
	cache    map[string]*PageMetadata
	expiry   map[string]time.Time
	ttl      time.Duration
	mutex    sync.RWMutex
	extractC chan extractRequest
}


type extractRequest struct {
	content  string
	filePath string
	result   chan *PageMetadata
}


var metadataCache = NewMetadataCache(8, 30*time.Minute)


func NewMetadataCache(workers int, ttl time.Duration) *MetadataCache {
	mc := &MetadataCache{
		cache:    make(map[string]*PageMetadata),
		expiry:   make(map[string]time.Time),
		ttl:      ttl,
		extractC: make(chan extractRequest, workers*2),
	}

	for i := 0; i < workers; i++ {
		go mc.worker()
	}

	go mc.cleanExpired(30 * time.Minute)

	return mc
}


func (mc *MetadataCache) worker() {
	for req := range mc.extractC {
		metadata := extractPageMetadataInternal(req.content, req.filePath)
		req.result <- metadata
	}
}


func (mc *MetadataCache) cleanExpired(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		mc.mutex.Lock()
		for key, expiry := range mc.expiry {
			if now.After(expiry) {
				delete(mc.cache, key)
				delete(mc.expiry, key)
			}
		}
		mc.mutex.Unlock()
	}
}


func (mc *MetadataCache) Get(key string) (*PageMetadata, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	metadata, ok := mc.cache[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(mc.expiry[key]) {
		return nil, false
	}

	return metadata, true
}


func (mc *MetadataCache) Set(key string, metadata *PageMetadata) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.cache[key] = metadata
	mc.expiry[key] = time.Now().Add(mc.ttl)
}
