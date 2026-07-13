package swcr

import (
	"bufio"
	"log/slog"
	"os"
	"strings"
)

// CodeWriter reads source files, filters out blank and comment lines, and
// writes the remaining lines as formatted paragraphs into a Document.
type CodeWriter struct {
	CommentChars []string
	opts         *WriterOpts
	doc          *Document
}

// NewCodeWriter creates a CodeWriter. commentChars are the line prefixes that
// identify a comment (e.g. "#", "//"). commentChars defaults to ["#", "//"]
// when empty, matching the Python swcr behaviour.
func NewCodeWriter(
	commentChars []string,
	opts *WriterOpts,
	doc *Document,
) *CodeWriter {
	if len(commentChars) == 0 {
		commentChars = []string{"#", "//"}
	}
	return &CodeWriter{
		CommentChars: commentChars,
		opts:         opts,
		doc:          doc,
	}
}

// isBlankLine reports whether line is empty or whitespace-only.
func isBlankLine(line string) bool {
	return strings.TrimSpace(line) == ""
}

// isCommentLine reports whether line (after stripping leading whitespace)
// starts with one of the registered comment prefixes.
func (w *CodeWriter) isCommentLine(line string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	for _, c := range w.CommentChars {
		if strings.HasPrefix(trimmed, c) {
			return true
		}
	}
	return false
}

// WriteFiles reads each file in files, filters lines, and adds them to the
// underlying Document.
func (w *CodeWriter) WriteFiles(files []string) error {
	for _, path := range files {
		if err := w.writeFile(path); err != nil {
			return err
		}
	}
	return nil
}

func (w *CodeWriter) writeFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), " \t\r")
		if isBlankLine(line) {
			continue
		}
		if w.isCommentLine(line) {
			continue
		}
		w.doc.AddParagraph(line)
	}
	if err := sc.Err(); err != nil {
		return err
	}
	slog.Debug("processed file", "path", path)
	return nil
}

// WriteHeader sets the document header.
func (w *CodeWriter) WriteHeader(title string) {
	w.doc.SetHeader(title)
}

// Save writes the completed .docx to disk.
func (w *CodeWriter) Save() error {
	return w.doc.Save()
}
