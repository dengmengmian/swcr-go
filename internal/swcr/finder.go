// Package swcr implements the core logic for the SWCR (Software Copyright
// Registration) source code document generator.
package swcr

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// autoExcludeDirs are directory names that are automatically skipped unless
// auto-exclude is disabled. These are common build-output, dependency, and
// IDE directories that should never appear in software copyright materials.
var autoExcludeDirs = map[string]bool{
	"node_modules":     true,
	"vendor":           true,
	"__pycache__":      true,
	".venv":            true,
	"venv":             true,
	".tox":             true,
	"dist":             true,
	"build":            true,
	"target":           true,
	".next":            true,
	".nuxt":            true,
	".cache":           true,
	"bower_components": true,
	".idea":            true,
	".vscode":          true,
}

// autoExcludeExts are file extensions that are automatically skipped.
var autoExcludeExts = map[string]bool{
	".pyc":   true,
	".pyo":   true,
	".class": true,
	".o":     true,
	".so":    true,
	".dylib": true,
	".dll":   true,
	".exe":   true,
	".bin":   true,
	".obj":   true,
	".a":     true,
	".lib":   true,
	".wasm":  true,
}

// isAutoExcludedName reports whether the file/dir name matches an auto-exclude
// pattern (extension or minified file suffix).
func isAutoExcludedName(name string) bool {
	for ext := range autoExcludeExts {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	// Minified frontend assets.
	if strings.HasSuffix(name, ".min.js") || strings.HasSuffix(name, ".min.css") {
		return true
	}
	return false
}

// CodeFinder recursively scans directories for source code files matching
// given extensions, skipping hidden files/directories and excluded paths.
type CodeFinder struct {
	// Exts is the list of file suffixes to match (without leading dot, e.g. "py", "go").
	Exts []string

	// AutoExclude controls whether common build/dependency directories and
	// binary file extensions are automatically skipped. Defaults to true.
	AutoExclude bool
}

// NewCodeFinder creates a CodeFinder with the given extensions. If exts is
// empty, it defaults to ["py"] matching the original Python swcr behaviour.
func NewCodeFinder(exts []string) *CodeFinder {
	if len(exts) == 0 {
		exts = []string{"py"}
	}
	return &CodeFinder{Exts: exts, AutoExclude: true}
}

// isCode reports whether the filename (not full path) ends with one of the
// registered extensions.
func (f *CodeFinder) isCode(name string) bool {
	for _, ext := range f.Exts {
		if strings.HasSuffix(name, "."+ext) {
			return true
		}
	}
	return false
}

// isHidden reports whether the filename represents a hidden file or directory
// (name starts with '.').
func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

// shouldBeExcluded reports whether the absolute path should be skipped because
// it starts with one of the exclude prefix patterns.
func shouldBeExcluded(absPath string, excludes []string) bool {
	for _, ex := range excludes {
		if strings.HasPrefix(absPath, ex) {
			return true
		}
	}
	return false
}

// Find recursively scans indir for source code files, returning their absolute
// paths. It skips hidden entries and excluded paths.
func (f *CodeFinder) Find(indir string, excludes []string) ([]string, error) {
	absDir, err := filepath.Abs(indir)
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.WalkDir(absDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()

		// Skip hidden files/directories (like .git).
		if isHidden(name) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Auto-exclude common build/dependency directories and binary files.
		if f.AutoExclude {
			if d.IsDir() {
				if autoExcludeDirs[name] {
					slog.Debug("auto-excluding directory", "name", name, "path", path)
					return filepath.SkipDir
				}
			} else {
				if isAutoExcludedName(name) {
					slog.Debug("auto-excluding file", "name", name, "path", path)
					return nil
				}
			}
		}

		// Get absolute path for exclusion check.
		absPath := path
		if !filepath.IsAbs(path) {
			absPath, _ = filepath.Abs(path)
		}

		// Skip excluded paths.
		if shouldBeExcluded(absPath, excludes) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() && f.isCode(name) {
			files = append(files, absPath)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	slog.Debug("found code files", "dir", absDir, "count", len(files))
	return files, nil
}
