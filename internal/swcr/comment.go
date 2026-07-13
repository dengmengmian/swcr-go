package swcr

import "strings"

// blockCommentPair defines a block comment with distinct or identical open/close
// delimiters (e.g. {"/*", "*/"}, {"\"\"\"", "\"\"\""}).
type blockCommentPair struct {
	open  string
	close string
}

// CommentStripper removes blank lines, line-comment lines, and block comments
// from source code. It maintains internal state for multi-line block comments
// and must be reset between files.
type CommentStripper struct {
	lineChars  []string
	blockPairs []blockCommentPair
	inBlock    string // close delimiter when inside a block comment, "" otherwise
}

// NewCommentStripper creates a CommentStripper. lineChars are the line-comment
// prefixes (e.g. "#", "//"). blockPairs are OPEN:CLOSE pairs (e.g. "/*:*/").
// Both may be empty; in that case only blank lines are removed.
func NewCommentStripper(lineChars []string, blockPairSpecs []string) *CommentStripper {
	cs := &CommentStripper{
		lineChars: lineChars,
	}
	for _, spec := range blockPairSpecs {
		parts := strings.SplitN(spec, ":", 2)
		if len(parts) == 2 {
			cs.blockPairs = append(cs.blockPairs, blockCommentPair{
				open: parts[0], close: parts[1],
			})
		}
	}
	return cs
}

// DefaultBlockCommentPairs returns the recommended block-comment pairs
// covering the most common languages.
func DefaultBlockCommentPairs() []string {
	return []string{
		"/*:*/",         // C, C++, Java, Go, JS, TS, Rust, Swift, Kotlin, etc.
		"<!--:-->",      // HTML, XML
		"\"\"\":\"\"\"", // Python docstrings (double-quote)
		"''':'''",       // Python docstrings (single-quote)
	}
}

// Reset clears the internal block-comment state. Call between files.
func (cs *CommentStripper) Reset() {
	cs.inBlock = ""
}

// ProcessLine filters a single raw source line. It returns the line to include
// and a boolean indicating whether it should be kept.
//
// Rules (in order):
//  1. Inside a block comment → skip until close delimiter found.
//  2. Line starts with a block-comment opener (after whitespace) → enter block.
//  3. Blank line → skip.
//  4. Line starts with a line-comment char (after whitespace) → skip.
func (cs *CommentStripper) ProcessLine(rawLine string) (string, bool) {
	line := strings.TrimRight(rawLine, " \t\r")
	trimmed := strings.TrimLeft(line, " \t")

	// ── 1. Inside a block comment ──────────────────────────────────────────
	if cs.inBlock != "" {
		closeLen := len(cs.inBlock) // capture before clearing
		idx := strings.Index(line, cs.inBlock)
		if idx >= 0 {
			cs.inBlock = ""
			after := strings.TrimSpace(line[idx+closeLen:])
			// If there is code after the close delimiter, reconsider this line.
			if after != "" {
				return cs.processLineAfterBlock(after)
			}
		}
		return "", false
	}

	// ── 2. Block comment opener at line start ─────────────────────────────
	for _, bp := range cs.blockPairs {
		if !strings.HasPrefix(trimmed, bp.open) {
			continue
		}
		rest := trimmed[len(bp.open):]
		if closeIdx := strings.Index(rest, bp.close); closeIdx >= 0 {
			// Single-line block comment — check for code after the close.
			after := strings.TrimSpace(rest[closeIdx+len(bp.close):])
			if after != "" {
				return cs.processLineAfterBlock(after)
			}
			return "", false
		}
		// Multi-line block comment starts here.
		cs.inBlock = bp.close
		return "", false
	}

	// ── 3. Blank line ─────────────────────────────────────────────────────
	if trimmed == "" {
		return "", false
	}

	// ── 4. Line comment ───────────────────────────────────────────────────
	for _, lc := range cs.lineChars {
		if strings.HasPrefix(trimmed, lc) {
			return "", false
		}
	}

	return line, true
}

// processLineAfterBlock re-evaluates content that appears after a block-comment
// close on the same line.
func (cs *CommentStripper) processLineAfterBlock(text string) (string, bool) {
	line := strings.TrimRight(text, " \t\r")
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return "", false
	}
	for _, lc := range cs.lineChars {
		if strings.HasPrefix(trimmed, lc) {
			return "", false
		}
	}
	return line, true
}
