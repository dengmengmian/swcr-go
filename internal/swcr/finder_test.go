package swcr

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func makeTempTree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if content != "" {
			if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
		} else {
			// Create directory.
			if err := os.MkdirAll(full, 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
	return dir
}

func sortedBaseNames(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Base(p)
	}
	sort.Strings(out)
	return out
}

func TestFinder_DefaultExts(t *testing.T) {
	root := makeTempTree(t, map[string]string{
		"a.py":  "print(1)",
		"b.txt": "hello",
	})
	f := NewCodeFinder(nil)
	files, err := f.Find(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || filepath.Base(files[0]) != "a.py" {
		t.Errorf("expected only a.py, got %v", sortedBaseNames(files))
	}
}

func TestFinder_CustomExts(t *testing.T) {
	root := makeTempTree(t, map[string]string{
		"a.go":   "package main",
		"a.py":   "print(1)",
		"a.html": "<html>",
		"a.txt":  "text",
	})
	// Match .go and .html only.
	f := NewCodeFinder([]string{"go", "html"})
	files, err := f.Find(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	names := sortedBaseNames(files)
	if len(names) != 2 || names[0] != "a.go" || names[1] != "a.html" {
		t.Errorf("expected a.go and a.html, got %v", names)
	}
}

func TestFinder_SkipsHiddenFiles(t *testing.T) {
	root := makeTempTree(t, map[string]string{
		".hidden.py": "secret",
		"visible.py": "public",
	})
	f := NewCodeFinder([]string{"py"})
	files, err := f.Find(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	names := sortedBaseNames(files)
	if len(names) != 1 || names[0] != "visible.py" {
		t.Errorf("expected only visible.py, got %v", names)
	}
}

func TestFinder_SkipsHiddenDirs(t *testing.T) {
	root := makeTempTree(t, map[string]string{
		".hidden_dir/a.py": "nope",
		"src/main.py":      "yes",
	})
	f := NewCodeFinder([]string{"py"})
	files, err := f.Find(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	names := sortedBaseNames(files)
	if len(names) != 1 || names[0] != "main.py" {
		t.Errorf("expected only main.py, got %v", names)
	}
}

func TestFinder_ExcludesPaths(t *testing.T) {
	root := makeTempTree(t, map[string]string{
		"src/main.py":    "yes",
		"vendor/lib.py":  "no",
		"docs/readme.py": "no",
	})
	f := NewCodeFinder([]string{"py"})
	excludes := []string{
		filepath.Join(root, "vendor"),
		filepath.Join(root, "docs"),
	}
	files, err := f.Find(root, excludes)
	if err != nil {
		t.Fatal(err)
	}
	names := sortedBaseNames(files)
	if len(names) != 1 || names[0] != "main.py" {
		t.Errorf("expected only main.py, got %v", names)
	}
}

func TestFinder_Recursive(t *testing.T) {
	root := makeTempTree(t, map[string]string{
		"a.py":          "1",
		"sub/b.py":      "2",
		"sub/deep/c.py": "3",
	})
	f := NewCodeFinder([]string{"py"})
	files, err := f.Find(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d: %v", len(files), files)
	}
}

func TestFinder_EmptyDir(t *testing.T) {
	root := makeTempTree(t, nil)
	f := NewCodeFinder([]string{"py"})
	files, err := f.Find(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %v", files)
	}
}

func TestFinder_IsCodeChecksSuffixWithDot(t *testing.T) {
	// The extension "py" should match "file.py" but NOT "filepy" or "file.pyc".
	root := makeTempTree(t, map[string]string{
		"file.py":  "good",
		"file.pyc": "bad",
		"filepy":   "bad",
	})
	f := NewCodeFinder([]string{"py"})
	files, err := f.Find(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	names := sortedBaseNames(files)
	if len(names) != 1 || names[0] != "file.py" {
		t.Errorf("expected only file.py, got %v", names)
	}
}

func TestFinder_IsHidden(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{".hidden", true},
		{"visible", false},
		{".", true},
		{"..", true},
		{"", false},
		{".git", true},
	}
	for _, tt := range tests {
		if got := isHidden(tt.name); got != tt.expected {
			t.Errorf("isHidden(%q) = %v, want %v", tt.name, got, tt.expected)
		}
	}
}

func TestFinder_ShouldBeExcluded(t *testing.T) {
	excludes := []string{"/home/user/vendor", "/home/user/tests"}
	if !shouldBeExcluded("/home/user/vendor/lib.py", excludes) {
		t.Error("should be excluded")
	}
	if !shouldBeExcluded("/home/user/tests/test_main.py", excludes) {
		t.Error("should be excluded")
	}
	if shouldBeExcluded("/home/user/src/main.py", excludes) {
		t.Error("should NOT be excluded")
	}
}
