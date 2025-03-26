package header

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/copygen/copygen/internal/config"
	"github.com/copygen/copygen/internal/view"
	"github.com/fatih/color"
)

var (
	// commentPrefixes defines the comment style for each file type
	commentPrefixes = map[string]string{
		".go": "//",
	}
)

// Processor handles file header operations with thread-safe caching.
type Processor struct {
	cfg    *config.Config
	view   view.Renderer
	dryRun bool

	dir       string
	cache     map[string][]string
	cacheMu   sync.RWMutex
	exclude   []string
	excludeMu sync.RWMutex
}

// NewProcessor creates a new header processor.
func NewProcessor(cfg *config.Config, dir string, v view.Renderer, dryRun bool) *Processor {
	return &Processor{
		cfg:     cfg,
		dir:     dir,
		view:    v,
		dryRun:  dryRun,
		cache:   make(map[string][]string),
		exclude: cfg.Exclude,
	}
}

// Process processes all files in the directory, adding headers where needed.
func (p *Processor) Process() error {
	// Warm up the cache for small configs
	if len(p.cfg.Header) < 1024 {
		p.warmCache()
	}

	var processed int
	err := filepath.Walk(p.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		prefix, ok := commentPrefixes[ext]
		if !ok || p.isExcluded(path) {
			return nil
		}

		hasHeader, err := p.checkHeader(path, prefix)
		if err != nil {
			return fmt.Errorf("check header: %v", err)
		}

		if hasHeader {
			return nil
		}

		processed++
		if p.dryRun {
			p.view.Render(color.BlueString(fmt.Sprintf("Would add header to \"%s\"\n", filepath.Clean(path))))
			return nil
		}

		if err := p.addHeader(path, prefix); err != nil {
			return fmt.Errorf("add header: %v", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk dir: %v", err)
	}

	return nil
}

// addHeader writes the header to a file using atomic replacement.
func (p *Processor) addHeader(path, prefix string) error {
	header := p.getHeader(prefix)
	headerBytes := []byte(strings.Join(header, "\n") + "\n\n")

	// Open original file
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create temp file in same directory
	tmp, err := os.CreateTemp(filepath.Dir(path), "*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	// Cleanup temp file on error
	defer func() {
		if tmp != nil {
			tmp.Close()
			os.Remove(tmpPath)
		}
	}()

	// Write header first
	if _, err := tmp.Write(headerBytes); err != nil {
		return err
	}

	// Copy original content
	if _, err := io.Copy(tmp, f); err != nil {
		return err
	}

	// Preserve original permissions
	if stat, err := f.Stat(); err == nil {
		tmp.Chmod(stat.Mode())
	}

	// Close before rename
	if err := tmp.Close(); err != nil {
		return err
	}
	tmp = nil // Prevent defer cleanup

	// Atomic replace
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	p.view.Render(color.BlueString(fmt.Sprintf("Added header to \"%s\"\n", filepath.Clean(path))))
	return nil
}

// getHeader returns the cached or generated header.
func (p *Processor) getHeader(prefix string) []string {
	p.cacheMu.RLock()
	cached, ok := p.cache[prefix]
	p.cacheMu.RUnlock()

	if ok {
		return cached
	}

	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	// Check again in case another goroutine updated it
	if cached, ok := p.cache[prefix]; ok {
		return cached
	}

	header := p.buildHeader(prefix)
	p.cache[prefix] = header
	return header
}

// isExcluded checks if path matches any exclude pattern.
func (p *Processor) isExcluded(path string) bool {
	normPath := filepath.ToSlash(path)

	p.excludeMu.RLock()
	defer p.excludeMu.RUnlock()

	for _, pattern := range p.exclude {
		matched, err := filepath.Match(pattern, normPath)
		if err != nil {
			continue // Skip invalid patterns
		}
		if matched {
			return true
		}

		// Handle directory patterns
		if strings.Contains(pattern, "**") || strings.HasSuffix(pattern, "/") {
			dirPattern := strings.TrimSuffix(pattern, "/") + "/"
			if strings.HasPrefix(normPath+"/", dirPattern) {
				return true
			}
		}
	}
	return false
}

// checkHeader verifies if file has the expected header.
func (p *Processor) checkHeader(path, prefix string) (bool, error) {
	if p.cfg.Header == "" {
		return true, nil
	}

	header := p.getHeader(prefix)
	if len(header) == 0 {
		return true, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for _, line := range header {
		if !scanner.Scan() {
			return false, scanner.Err()
		}
		if scanner.Text() != line {
			return false, nil
		}
	}

	return true, nil
}

// buildHeader creates the formatted header lines.
func (p *Processor) buildHeader(prefix string) []string {
	lines := strings.Split(p.cfg.Header, "\n")
	var header []string

	for _, line := range lines {
		if line == "" {
			header = append(header, prefix)
		} else {
			header = append(header, prefix+" "+line)
		}
	}

	// Trim trailing empty comments
	for len(header) > 0 && header[len(header)-1] == prefix {
		header = header[:len(header)-1]
	}

	return header
}

// warmCache pre-generates headers for known file types.
func (p *Processor) warmCache() {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	for _, prefix := range commentPrefixes {
		if _, exists := p.cache[prefix]; !exists {
			p.cache[prefix] = p.buildHeader(prefix)
		}
	}
}
