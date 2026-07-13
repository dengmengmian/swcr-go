package swcr

import (
	"bufio"
	"log/slog"
	"os"
)

// CodeWriter reads source files, filters out blank and comment lines, and
// writes the remaining lines as formatted paragraphs into a Document.
type CodeWriter struct {
	stripper *CommentStripper
	opts     *WriterOpts
	doc      *Document
}

// NewCodeWriter creates a CodeWriter.
func NewCodeWriter(
	stripper *CommentStripper,
	opts *WriterOpts,
	doc *Document,
) *CodeWriter {
	return &CodeWriter{
		stripper: stripper,
		opts:     opts,
		doc:      doc,
	}
}

// CollectLines reads all files, applies the comment stripper, and returns
// every kept line in order (right-trimmed, indentation preserved).
func (w *CodeWriter) CollectLines(files []string) ([]string, error) {
	w.stripper.Reset()
	var all []string
	for _, path := range files {
		lines, err := w.collectFile(path)
		if err != nil {
			return nil, err
		}
		all = append(all, lines...)
	}
	return all, nil
}

func (w *CodeWriter) collectFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var kept []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line, ok := w.stripper.ProcessLine(sc.Text()); ok {
			kept = append(kept, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	slog.Debug("processed file", "path", path, "kept", len(kept))
	return kept, nil
}

// WriteLines writes the given lines to the underlying Document as formatted
// paragraphs.
func (w *CodeWriter) WriteLines(lines []string) {
	for _, line := range lines {
		w.doc.AddParagraph(line)
	}
}

// WriteFiles reads each file, filters lines through the comment stripper,
// and writes them to the underlying Document. It resets the stripper state
// before processing.
func (w *CodeWriter) WriteFiles(files []string) error {
	w.stripper.Reset()
	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if line, ok := w.stripper.ProcessLine(sc.Text()); ok {
				w.doc.AddParagraph(line)
			}
		}
		f.Close()
		if err := sc.Err(); err != nil {
			return err
		}
		slog.Debug("processed file", "path", path)
	}
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
