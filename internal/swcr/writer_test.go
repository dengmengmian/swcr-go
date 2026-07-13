package swcr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	if err := os.WriteFile(src, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "out.docx")
	opts := DefaultWriterOpts()
	doc, err := NewDocument(out, opts)
	if err != nil {
		t.Fatal(err)
	}

	stripper := NewCommentStripper([]string{"#", "//"}, nil)
	w := NewCodeWriter(stripper, opts, doc)
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

func TestCodeWriter_CollectLines(t *testing.T) {
	dir := t.TempDir()

	// File 1: Python
	src1 := filepath.Join(dir, "a.py")
	if err := os.WriteFile(src1, []byte("# comment\nprint(1)\n\nprint(2)\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// File 2: Go
	src2 := filepath.Join(dir, "b.go")
	if err := os.WriteFile(src2, []byte("// comment\npackage main\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	stripper := NewCommentStripper([]string{"#", "//"}, nil)
	w := NewCodeWriter(stripper, DefaultWriterOpts(), nil)
	lines, err := w.CollectLines([]string{src1, src2})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"print(1)", "print(2)", "package main"}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d: %v", len(lines), len(want), lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d: got %q, want %q", i, lines[i], want[i])
		}
	}
}

func TestCodeWriter_BlockCommentInFiles(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "test.c")
	content := "int main() {\n    /* block\n       comment */\n    return 0;\n}\n"
	if err := os.WriteFile(src, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	stripper := NewCommentStripper([]string{"//"}, []string{"/*:*/"})
	w := NewCodeWriter(stripper, DefaultWriterOpts(), nil)
	lines, err := w.CollectLines([]string{src})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"int main() {", "    return 0;", "}"}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d: %v", len(lines), len(want), lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d: got %q, want %q", i, lines[i], want[i])
		}
	}
}
