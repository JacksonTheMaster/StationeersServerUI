package backupmgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	defaultAnalysisConcurrency = 1
	defaultAnalysisCacheSize   = 256
)

type saveIdentity struct {
	size       int64
	modifiedNS int64
}

type cachedSaveAnalysis struct {
	identity saveIdentity
	value    SaveAnalysis
	access   uint64
}

// SaveAnalyzer bounds expensive deep scans and caches immutable backup results.
// Cache entries are invalidated automatically if a file's size or modification
// time changes.
type SaveAnalyzer struct {
	semaphore chan struct{}
	maxCached int
	analyze   func(context.Context, string) (SaveAnalysis, error)

	mu      sync.Mutex
	clock   uint64
	entries map[string]cachedSaveAnalysis
}

func NewSaveAnalyzer(maxConcurrent, maxCached int) *SaveAnalyzer {
	if maxConcurrent <= 0 {
		maxConcurrent = defaultAnalysisConcurrency
	}
	if maxCached < 0 {
		maxCached = defaultAnalysisCacheSize
	}

	return &SaveAnalyzer{
		semaphore: make(chan struct{}, maxConcurrent),
		maxCached: maxCached,
		analyze:   AnalyzeSave,
		entries:   make(map[string]cachedSaveAnalysis),
	}
}

func (analyzer *SaveAnalyzer) Analyze(ctx context.Context, path string) (SaveAnalysis, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	canonicalPath, err := filepath.Abs(path)
	if err != nil {
		return SaveAnalysis{}, fmt.Errorf("resolve save path: %w", err)
	}

	select {
	case analyzer.semaphore <- struct{}{}:
		defer func() { <-analyzer.semaphore }()
	case <-ctx.Done():
		return SaveAnalysis{}, ctx.Err()
	}

	before, err := identifySave(canonicalPath)
	if err != nil {
		return SaveAnalysis{}, err
	}
	if cached, ok := analyzer.cached(canonicalPath, before); ok {
		return cached, nil
	}

	analysis, err := analyzer.analyze(ctx, canonicalPath)
	if err != nil {
		return SaveAnalysis{}, err
	}
	after, err := identifySave(canonicalPath)
	if err != nil {
		return SaveAnalysis{}, err
	}
	if before != after {
		return SaveAnalysis{}, fmt.Errorf("save archive changed during analysis")
	}

	analyzer.store(canonicalPath, after, analysis)
	return analysis, nil
}

func (analyzer *SaveAnalyzer) Invalidate(path string) {
	canonicalPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	analyzer.mu.Lock()
	delete(analyzer.entries, canonicalPath)
	analyzer.mu.Unlock()
}

func (analyzer *SaveAnalyzer) cached(path string, identity saveIdentity) (SaveAnalysis, bool) {
	if analyzer.maxCached == 0 {
		return SaveAnalysis{}, false
	}

	analyzer.mu.Lock()
	defer analyzer.mu.Unlock()
	entry, ok := analyzer.entries[path]
	if !ok || entry.identity != identity {
		return SaveAnalysis{}, false
	}
	analyzer.clock++
	entry.access = analyzer.clock
	analyzer.entries[path] = entry
	return entry.value, true
}

func (analyzer *SaveAnalyzer) store(path string, identity saveIdentity, value SaveAnalysis) {
	if analyzer.maxCached == 0 {
		return
	}

	analyzer.mu.Lock()
	defer analyzer.mu.Unlock()
	if _, exists := analyzer.entries[path]; !exists && len(analyzer.entries) >= analyzer.maxCached {
		analyzer.evictOldest()
	}
	analyzer.clock++
	analyzer.entries[path] = cachedSaveAnalysis{identity: identity, value: value, access: analyzer.clock}
}

// evictOldest must be called with analyzer.mu held.
func (analyzer *SaveAnalyzer) evictOldest() {
	var oldestPath string
	var oldestAccess uint64
	first := true
	for path, entry := range analyzer.entries {
		if first || entry.access < oldestAccess {
			oldestPath = path
			oldestAccess = entry.access
			first = false
		}
	}
	if !first {
		delete(analyzer.entries, oldestPath)
	}
}

func identifySave(path string) (saveIdentity, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return saveIdentity{}, fmt.Errorf("stat save archive: %w", err)
	}
	if !stat.Mode().IsRegular() {
		return saveIdentity{}, fmt.Errorf("save archive is not a regular file")
	}
	return saveIdentity{size: stat.Size(), modifiedNS: stat.ModTime().UnixNano()}, nil
}
