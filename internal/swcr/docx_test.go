package swcr

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"path/filepath"
	"strings"
	"testing"
)

// readZipEntry reads the content of a named entry from a ZIP file.
func readZipEntry(t *testing.T, zipPath, entryName string) []byte {
	t.Helper()
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer r.Close()
	for _, f := range r.File {
		if f.Name == entryName {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open entry %s: %v", entryName, err)
			}
			defer rc.Close()
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(rc); err != nil {
				t.Fatalf("read entry %s: %v", entryName, err)
			}
			return buf.Bytes()
		}
	}
	t.Fatalf("entry %s not found in zip", entryName)
	return nil
}

// se is a simplified XML element for structural assertions.
// Namespace prefixes are stripped — only Local names are compared.
type se struct {
	Name     string
	Attrs    map[string]string
	Children []*se
	Text     string
}

// find recursively finds the first element matching local name.
func (e *se) find(name string) *se {
	if e.Name == name {
		return e
	}
	for _, c := range e.Children {
		if found := c.find(name); found != nil {
			return found
		}
	}
	return nil
}

// findAll recursively finds all elements matching local name.
func (e *se) findAll(name string) []*se {
	var out []*se
	if e.Name == name {
		out = append(out, e)
	}
	for _, c := range e.Children {
		out = append(out, c.findAll(name)...)
	}
	return out
}

// parseSE parses XML into a simplified element tree.
// Namespace prefixes are stripped — only Local names are compared.
func parseSE(t *testing.T, data []byte) *se {
	t.Helper()
	d := xml.NewDecoder(bytes.NewReader(data))
	var root *se
	var stack []*se
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			el := &se{
				Name:  t.Name.Local,
				Attrs: make(map[string]string),
			}
			for _, a := range t.Attr {
				el.Attrs[a.Name.Local] = a.Value
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, el)
			} else {
				root = el
			}
			stack = append(stack, el)
		case xml.CharData:
			if len(stack) > 0 {
				text := strings.TrimSpace(string(t))
				if text != "" {
					stack[len(stack)-1].Text += text
				}
			}
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return root
}

// TestDocx_AssertNoCommentOrBlankLines generates a DOCX with mixed content
// and verifies the resulting document.xml contains no comment lines or blank
// lines.
func TestDocx_AssertNoCommentOrBlankLines(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.docx")
	opts := DefaultWriterOpts()
	doc, err := NewDocument(tmp, opts)
	if err != nil {
		t.Fatal(err)
	}

	lines := []string{
		"package main",
		"",
		"// comment",
		"func main() {",
		"# python comment",
		"   ",
		"	println(\"hello\")",
		"	// inline but line-start comment",
		"	/* block start",
		"	fmt.Println(\"// not a comment\")",
	}
	for _, l := range lines {
		l = strings.TrimRight(l, " \t\r")
		if isBlankLine(l) {
			continue
		}
		cw := &CodeWriter{CommentChars: []string{"#", "//"}}
		if cw.isCommentLine(l) {
			continue
		}
		doc.AddParagraph(l)
	}
	doc.SetHeader("TestApp V1.0")
	if err := doc.Save(); err != nil {
		t.Fatal(err)
	}

	docXML := readZipEntry(t, tmp, "word/document.xml")
	root := parseSE(t, docXML)

	// Root must be "document"
	if root.Name != "document" {
		t.Fatalf("root element is %q, want %q", root.Name, "document")
	}

	// All w:t elements should not contain comment prefixes or be empty.
	texts := root.findAll("t")
	for _, tt := range texts {
		if tt.Text == "" {
			t.Error("blank line leaked: empty w:t")
		}
		trimmed := strings.TrimLeft(tt.Text, " \t")
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			t.Errorf("comment line leaked: %q", tt.Text)
		}
	}

	// Expected 5 code lines.
	if len(texts) != 5 {
		t.Errorf("expected 5 w:t elements, got %d", len(texts))
	}
}

// TestDocx_HeaderContainsTitleAndPageField verifies the header XML contains
// the title text and a PAGE field instruction.
func TestDocx_HeaderContainsTitleAndPageField(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test_header.docx")
	opts := DefaultWriterOpts()
	doc, err := NewDocument(tmp, opts)
	if err != nil {
		t.Fatal(err)
	}
	doc.SetHeader("MySoftware V2.0")
	doc.AddParagraph("code line")
	if err := doc.Save(); err != nil {
		t.Fatal(err)
	}

	hdrXML := readZipEntry(t, tmp, "word/header1.xml")
	root := parseSE(t, hdrXML)

	if root.Name != "hdr" {
		t.Fatalf("header root is %q, want %q", root.Name, "hdr")
	}

	foundTitle := false
	foundPageField := false
	for _, tt := range root.findAll("t") {
		if strings.Contains(tt.Text, "MySoftware V2.0") {
			foundTitle = true
		}
	}
	for _, it := range root.findAll("instrText") {
		if strings.Contains(it.Text, "PAGE") {
			foundPageField = true
		}
	}
	if !foundTitle {
		t.Error("header does not contain title 'MySoftware V2.0'")
	}
	if !foundPageField {
		t.Error("header does not contain PAGE field")
	}
}

// TestDocx_ParagraphProperties verifies the paragraph spacing and font
// properties match the settings for 50-lines-per-page.
func TestDocx_ParagraphProperties(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test_props.docx")
	opts := DefaultWriterOpts()
	doc, err := NewDocument(tmp, opts)
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("test line")
	if err := doc.Save(); err != nil {
		t.Fatal(err)
	}

	docXML := readZipEntry(t, tmp, "word/document.xml")
	root := parseSE(t, docXML)

	sp := root.find("spacing")
	if sp == nil {
		t.Fatal("no w:spacing element found")
	}
	if sp.Attrs["after"] != "46" {
		t.Errorf("space after: got %q, want 46 (2.3pt)", sp.Attrs["after"])
	}
	if sp.Attrs["line"] != "210" {
		t.Errorf("line spacing: got %q, want 210 (10.5pt)", sp.Attrs["line"])
	}
	if sp.Attrs["lineRule"] != "exact" {
		t.Errorf("line rule: got %q, want 'exact'", sp.Attrs["lineRule"])
	}
	// before=0 → attribute may be omitted
	if sp.Attrs["before"] != "" {
		t.Errorf("space before: got %q, want empty/omitted (0)", sp.Attrs["before"])
	}

	rf := root.find("rFonts")
	if rf == nil {
		t.Fatal("no w:rFonts element found")
	}
	if rf.Attrs["eastAsia"] != "宋体" {
		t.Errorf("eastAsia font: got %q, want 宋体", rf.Attrs["eastAsia"])
	}

	sz := root.find("sz")
	if sz == nil {
		t.Fatal("no w:sz element found")
	}
	if sz.Attrs["val"] != "21" {
		t.Errorf("font size: got %q half-pts, want 21 (10.5pt)", sz.Attrs["val"])
	}
}

// TestDocx_ValidZipStructure verifies essential entries exist in the ZIP.
func TestDocx_ValidZipStructure(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test_structure.docx")
	opts := DefaultWriterOpts()
	doc, err := NewDocument(tmp, opts)
	if err != nil {
		t.Fatal(err)
	}
	doc.SetHeader("Test")
	doc.AddParagraph("line")
	if err := doc.Save(); err != nil {
		t.Fatal(err)
	}

	r, err := zip.OpenReader(tmp)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer r.Close()

	required := map[string]bool{
		"[Content_Types].xml":          false,
		"_rels/.rels":                  false,
		"word/_rels/document.xml.rels": false,
		"word/document.xml":            false,
		"word/header1.xml":             false,
	}
	for _, f := range r.File {
		if _, ok := required[f.Name]; ok {
			required[f.Name] = true
		}
	}
	for name, found := range required {
		if !found {
			t.Errorf("missing required entry: %s", name)
		}
	}
}

// TestDocx_DocumentStructure verifies the correct OOXML structure:
// root = w:document, child = w:body, sectPr = last child of w:body.
func TestDocx_DocumentStructure(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test_struct.docx")
	opts := DefaultWriterOpts()
	doc, err := NewDocument(tmp, opts)
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("line 1")
	doc.AddParagraph("line 2")
	if err := doc.Save(); err != nil {
		t.Fatal(err)
	}

	docXML := readZipEntry(t, tmp, "word/document.xml")
	root := parseSE(t, docXML)

	// 1. Root must be "document"
	if root.Name != "document" {
		t.Fatalf("root element: got %q, want 'document'", root.Name)
	}

	// 2. Root must have a single child "body"
	bodies := root.findAll("body")
	if len(bodies) != 1 {
		t.Fatalf("expected exactly 1 w:body under w:document, got %d", len(bodies))
	}
	body := bodies[0]

	// 3. w:body must have at least one w:p and exactly one w:sectPr
	if len(body.findAll("p")) < 1 {
		t.Error("w:body has no w:p children")
	}
	sectPrs := body.findAll("sectPr")
	if len(sectPrs) != 1 {
		t.Fatalf("expected exactly 1 w:sectPr, got %d", len(sectPrs))
	}

	// 4. w:sectPr must be the LAST child of w:body
	lastChild := body.Children[len(body.Children)-1]
	if lastChild.Name != "sectPr" {
		t.Errorf("last child of w:body is %q, want 'sectPr'", lastChild.Name)
	}
}
