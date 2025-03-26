// Copyright (c) Copygen. Licensed under the Apache License, Version 2.0.
// See LICENSE for details. Do not modify this header – changes will be overwritten.

package header

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bschaatsbergen/copygen/internal/config"
	"github.com/bschaatsbergen/copygen/internal/view"
	"github.com/fatih/color"
)

var (
	// commentPrefixes defines the comment style for each file type
	commentPrefixes = map[string]string{
		".go": "//",
	}
)

// Processor handles file header operations.
type Processor struct {
	cfg    *config.Config
	view   view.Renderer
	dryRun bool
	dir    string

	exclude []string
}

// NewProcessor creates a new header processor.
func NewProcessor(cfg *config.Config, dir string, v view.Renderer, dryRun bool) *Processor {
	return &Processor{
		cfg:     cfg,
		dir:     dir,
		view:    v,
		dryRun:  dryRun,
		exclude: cfg.Exclude,
	}
}

// Update the Process method to use the new function name
func (p *Processor) Process() error {
	return filepath.Walk(p.dir, func(path string, info os.FileInfo, err error) error {
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

		hasHeader, err := p.peekHeader(path, prefix)
		if err != nil {
			return fmt.Errorf("peek header: %v", err)
		}

		if hasHeader {
			return nil
		}

		if p.dryRun {
			p.view.Render(color.BlueString(fmt.Sprintf("would add header to %s\n", filepath.Clean(path))))
			return nil
		}

		if err := p.addHeader(path, prefix); err != nil {
			return err
		}

		return nil
	})
}

// peekHeader checks if the file already has the expected header.
func (p *Processor) peekHeader(path, prefix string) (bool, error) {
	if p.cfg.Header == "" {
		return true, nil
	}

	expectedHeader := p.buildHeader(prefix)
	if len(expectedHeader) == 0 {
		return true, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Peek at the file's beginning lines to check for header
	scanner := bufio.NewScanner(f)
	for _, headerLine := range expectedHeader {
		if !scanner.Scan() {
			return false, scanner.Err()
		}
		if scanner.Text() != headerLine {
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

// addHeader writes the header to a file using atomic replacement.
func (p *Processor) addHeader(path, prefix string) error {
	header := p.buildHeader(prefix)
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

	p.view.Render(color.BlueString(fmt.Sprintf("added header to %s\n", filepath.Base(path))))
	return nil
}

// isExcluded checks if path matches any exclude pattern.
func (p *Processor) isExcluded(path string) bool {
	normPath := filepath.ToSlash(path)

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
