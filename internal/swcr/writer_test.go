package swcr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsBlankLine(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"", true},
		{"   ", true},
		{"\t\t", true},
		{"  \t  ", true},
		{"code", false},
		{"  code  ", false},
	}
	for _, tt := range tests {
		if got := isBlankLine(tt.line); got != tt.expected {
			t.Errorf("isBlankLine(%q) = %v, want %v", tt.line, got, tt.expected)
		}
	}
}

func TestIsCommentLine(t *testing.T) {
	w := &CodeWriter{CommentChars: []string{"#", "//"}}
	tests := []struct {
		line     string
		expected bool
	}{
		{"# comment", true},
		{"  # indented comment", true},
		{"// comment", true},
		{"  // indented comment", true},
		{"code # not comment", false},
		{"  code // still not", false},
		{"/* block comment", false},
		{"code", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := w.isCommentLine(tt.line); got != tt.expected {
			t.Errorf("isCommentLine(%q) = %v, want %v", tt.line, got, tt.expected)
		}
	}
}

func TestCodeWriter_FiltersBlanksAndComments(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "test.py")
	content := `# header comment

def hello():
    # inline comment
    return "hello"

// This is not a Python comment but would be caught

print("done")
`
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "out.docx")
	opts := DefaultWriterOpts()
	doc, err := NewDocument(out, opts)
	if err != nil {
		t.Fatal(err)
	}

	w := NewCodeWriter([]string{"#", "//"}, opts, doc)
	w.WriteHeader("Test")
	if err := w.WriteFiles([]string{src}); err != nil {
		t.Fatal(err)
	}
	if err := w.Save(); err != nil {
		t.Fatal(err)
	}

	docXML := readZipEntry(t, out, "word/document.xml")
	root := parseSE(t, docXML)

	texts := root.findAll("t")
	for _, tt := range texts {
		txt := tt.Text
		if txt == "" {
			t.Errorf("blank line leaked: empty text run")
		}
		trimmed := strings.TrimLeft(txt, " \t")
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			t.Errorf("comment line leaked: %q", txt)
		}
	}

	// Expect only: def hello():, return "hello", print("done")
	if len(texts) != 3 {
		t.Errorf("expected 3 code lines, got %d", len(texts))
	}
}
